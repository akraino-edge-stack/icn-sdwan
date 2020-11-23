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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var sdewanApplicationHandler = new(SdewanApplicationHandler)

type SdewanApplicationHandler struct {
}

func (m *SdewanApplicationHandler) GetType() string {
	return "sdewanApplication"
}

func (m *SdewanApplicationHandler) GetName(instance runtime.Object) string {
	service := instance.(*batchv1alpha1.SdewanApplication)
	return service.Name
}

func (m *SdewanApplicationHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *SdewanApplicationHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.SdewanApplication{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *SdewanApplicationHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	return nil, nil
}

func (m *SdewanApplicationHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	return false
}

func (m *SdewanApplicationHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	return nil, nil
}

func (m *SdewanApplicationHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	return nil, nil
}

func (m *SdewanApplicationHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	return nil, nil
}

func (m *SdewanApplicationHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	return nil
}

func (m *SdewanApplicationHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("cnfService", "restart")
}

// SdewanApplicationReconciler reconciles a SdewanApplication object
type SdewanApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=sdewanapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=sdewanapplications/status,verbs=get;update;patch

func (r *SdewanApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("sdewanapplication", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *SdewanApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.SdewanApplication{}).
		Watches(
			&source.Kind{Type: &corev1.Service{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetServiceToRequestsFunc(r)),
			},
			IPFilter).
		Complete(r)
}
