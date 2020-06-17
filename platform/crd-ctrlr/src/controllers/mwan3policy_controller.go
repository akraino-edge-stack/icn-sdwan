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

package controllers

import (
	"context"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var mwan3PolicyHandler = new(Mwan3PolicyHandler)

type Mwan3PolicyHandler struct {
}

func (m *Mwan3PolicyHandler) GetType() string {
	return "Mwan3Policy"
}

func (m *Mwan3PolicyHandler) GetName(instance runtime.Object) string {
	policy := instance.(*batchv1alpha1.Mwan3Policy)
	return policy.Name
}

func (m *Mwan3PolicyHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *Mwan3PolicyHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.Mwan3Policy{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *Mwan3PolicyHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	policy := instance.(*batchv1alpha1.Mwan3Policy)
	members := make([]openwrt.SdewanMember, len(policy.Spec.Members))
	for i, membercr := range policy.Spec.Members {
		iface, err := net2iface(membercr.Network, deployment)
		if err != nil {
			return nil, err
		}
		members[i] = openwrt.SdewanMember{
			Interface: iface,
			Metric:    strconv.Itoa(membercr.Metric),
			Weight:    strconv.Itoa(membercr.Weight),
		}
	}
	return &openwrt.SdewanPolicy{Name: policy.Name, Members: members}, nil
}

func (m *Mwan3PolicyHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	policy1 := instance1.(*openwrt.SdewanPolicy)
	policy2 := instance2.(*openwrt.SdewanPolicy)
	return reflect.DeepEqual(*policy1, *policy2)
}

func (m *Mwan3PolicyHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	ret, err := mwan3.GetPolicy(name)
	return ret, err
}

func (m *Mwan3PolicyHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	policy := instance.(*openwrt.SdewanPolicy)
	return mwan3.CreatePolicy(*policy)
}

func (m *Mwan3PolicyHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	policy := instance.(*openwrt.SdewanPolicy)
	return mwan3.UpdatePolicy(*policy)
}

func (m *Mwan3PolicyHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	return mwan3.DeletePolicy(name)
}

func (m *Mwan3PolicyHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("mwan3", "restart")
}

// Mwan3PolicyReconciler reconciles a Mwan3Policy object
type Mwan3PolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=mwan3policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=mwan3policies/status,verbs=get;update;patch

func (r *Mwan3PolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, mwan3PolicyHandler)
}

func (r *Mwan3PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.Mwan3Policy{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.Mwan3PolicyList{})),
			},
			Filter).
		Complete(r)
}
