/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/minio/minio/pkg/wildcard"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var bucketlog = logf.Log.WithName("sdewan-bucket-permission")

func (r *Mwan3Policy) SetupWebhookWithManager2(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().Register(
		"/validate-sdewan-bucket-permission",
		&webhook.Admission{Handler: &bucketPermissionValidator{Client: mgr.GetClient()}})
	return nil
}

// +kubebuilder:webhook:path=/validate-sdewan-bucket-permission,mutating=false,failurePolicy=fail,groups="batch.sdewan.akraino.org",resources=mwan3policies;mwan3rules,verbs=create;update;delete,versions=v1alpha1,name=validate-sdewan-bucket.akraino.org

// bucketPermissionValidator validates Pods
type bucketPermissionValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// map key is the resource type, values is the permissions. Sample bucket permission:
//   { "mwan3rules": ["app-intent", "k8s-service"], "mwan3policies": ["app-intent"] }
type BucketPermission map[string][]string

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;roles;rolebindings;clusterrolebindings,verbs=get;list;watch

// bucketPermissionValidator admits a pod iff a specific annotation exists.
func (v *bucketPermissionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Kind.Group != "batch.sdewan.akraino.org" {
		return admission.Errored(
			http.StatusBadRequest,
			errors.New("The group is not batch.sdewan.akraino.org"))
	}
	authenticated := false
	clusterAdmin := false
	for _, g := range req.UserInfo.Groups {
		if g == "system:masters" {
			clusterAdmin = true
		}
		if g == "system:authenticated" {
			authenticated = true
		}
	}
	if authenticated && clusterAdmin {
		return admission.Allowed("Allowd as cluster admin")
	}
	var meta metav1.ObjectMeta
	var err error
	var obj runtime.Object
	switch req.Kind.Kind {
	case "Mwan3Policy":
		obj = &Mwan3Policy{}
	case "Mwan3Rule":
		obj = &Mwan3Rule{}
	default:
		return admission.Errored(
			http.StatusBadRequest,
			errors.New(fmt.Sprintf("Kind is not supported: %v", req.Kind)))
	}

	if req.Operation == "CREATE" || req.Operation == "UPDATE" {
		err = v.decoder.Decode(req, obj)
	} else if req.Operation == "DELETE" {
		err = v.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name}, obj)
	} else {
		return admission.Denied(fmt.Sprintf("We don't support operation type: %s", req.Operation))
	}
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// objectmeta is the second field in Object, so Field(1)
	meta = reflect.ValueOf(obj).Elem().Field(1).Interface().(metav1.ObjectMeta)
	bucketType := meta.Labels["sdewan-bucket-type"]
	if bucketType == "" {
		return admission.Allowed("")
	}

	// Validate role within target namespace
	roleBindings := &rbacv1.RoleBindingList{}
	err = v.Client.List(ctx, roleBindings, client.MatchingFields{".subjects": req.UserInfo.Username})
	if err != nil {
		bucketlog.Error(err, "Failed to get rolebinding list")
	} else {
		for _, rolebinding := range roleBindings.Items {
			role := &rbacv1.Role{}
			err = v.Client.Get(ctx, types.NamespacedName{Namespace: rolebinding.Namespace, Name: rolebinding.RoleRef.Name}, role)
			if err != nil {
				bucketlog.Error(err, "Failed to get role from rolebinding")
				continue
			}
			if role.Annotations["sdewan-bucket-type-permission"] == "" {
				continue
			}
			var perm BucketPermission = make(map[string][]string)
			err := json.Unmarshal([]byte(role.Annotations["sdewan-bucket-type-permission"]), &perm)
			if err != nil {
				bucketlog.Error(err, "Failed to parse bucket permission annotation")
				continue
			}
			if role.Namespace == meta.Namespace {
				for res, resPerm := range perm {
					if wildcard.Match(res, req.Resource.Resource) {
						for _, p := range resPerm {
							if wildcard.Match(p, bucketType) {
								return admission.Allowed("")
							}
						}
					}
				}
			}
		}
	}
	// Validate clusterrole
	clusterRoleBindings := &rbacv1.ClusterRoleBindingList{}
	err = v.Client.List(ctx, clusterRoleBindings, client.MatchingFields{".subjects": req.UserInfo.Username})
	if err != nil {
		bucketlog.Error(err, "Failed to get clusterrolebinding list")
	} else {
		for _, clusterrolebinding := range clusterRoleBindings.Items {
			clusterrole := &rbacv1.ClusterRole{}
			err = v.Client.Get(ctx, types.NamespacedName{Name: clusterrolebinding.RoleRef.Name}, clusterrole)
			if err != nil {
				bucketlog.Error(err, "Failed to get clusterrole from clusterrolebinding")
				continue
			}
			if clusterrole.Annotations["sdewan-bucket-type-permission"] == "" {
				continue
			}
			var perm BucketPermission = make(map[string][]string)
			err := json.Unmarshal([]byte(clusterrole.Annotations["sdewan-bucket-type-permission"]), &perm)
			if err != nil {
				bucketlog.Error(err, "Failed to parse bucket permission annotation")
				continue
			}
			for res, resPerm := range perm {
				if wildcard.Match(res, req.Resource.Resource) {
					for _, p := range resPerm {
						if wildcard.Match(p, bucketType) {
							return admission.Allowed("")
						}
					}
				}
			}
		}
	}

	return admission.Denied(fmt.Sprintf("User(%v) don't have the permission", req.UserInfo.Username))
}

// bucketPermissionValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *bucketPermissionValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
