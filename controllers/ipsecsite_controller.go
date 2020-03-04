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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
)

// IpsecSiteReconciler reconciles a IpsecSite object
type IpsecSiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecsites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecsites/status,verbs=get;update;patch

func (r *IpsecSiteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ipsecsite", req.NamespacedName)

	// your logic here
	instance := &batchv1alpha1.IpsecSite{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	foundSdewans := &batchv1alpha1.SdewanList{}
	err = r.List(ctx, foundSdewans, client.MatchingFields{".spec.IPSecSite": instance.Name})
	if err != nil && errors.IsNotFound(err) {
		log.Info("No sdewan using this ipsec conf", "ipsecsite", instance.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get the sdewan list using this ipsec conf", "ipsecsite", instance.Name)
		return ctrl.Result{}, nil
	} else {
		log.Info("Applying ipsec configurations for selected sdewan instances as the conf changes", "ipsecsite", instance.Name)
		for _, sdewan := range foundSdewans.Items {
			// Updating sdewan to set status isapplied = false
			// this will trigger sdewan controller to apply the new conf
			sdewan.Status.IpsecStatus = batchv1alpha1.IPSecStatus{Name: instance.Name, IsApplied: false}
			err := r.Status().Update(ctx, &sdewan)
			if err != nil {
				log.Error(err, "Failed to update the sdewan instance status", "sdewan", sdewan.Name)
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *IpsecSiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&batchv1alpha1.Sdewan{}, ".spec.IPSecSite", func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		sdewan := rawObj.(*batchv1alpha1.Sdewan)
		return []string{sdewan.Spec.IpsecSite}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.IpsecSite{}).
		Complete(r)
}
