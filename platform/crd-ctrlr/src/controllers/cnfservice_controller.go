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

var cnfServiceHandler = new(CNFServiceHandler)

type CNFServiceHandler struct {
}

func (m *CNFServiceHandler) GetType() string {
	return "cnfService"
}

func (m *CNFServiceHandler) GetName(instance runtime.Object) string {
	service := instance.(*batchv1alpha1.CNFService)
	return service.Name
}

func (m *CNFServiceHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *CNFServiceHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.CNFService{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *CNFServiceHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	svc := instance.(*batchv1alpha1.CNFService)
	openwrtsvc := openwrt.SdewanSvc{
		Name:     svc.Name,
		FullName: svc.Spec.FullName,
		Port:     svc.Spec.Port,
		DPort:    svc.Spec.DPort,
	}
	return &openwrtsvc, nil
}

func (m *CNFServiceHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	service1 := instance1.(*openwrt.SdewanSvc)
	service2 := instance2.(*openwrt.SdewanSvc)
	return reflect.DeepEqual(*service1, *service2)
}

func (m *CNFServiceHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	svc := openwrt.SvcClient{OpenwrtClient: openwrtClient}
	ret, err := svc.GetSvc(name)
	return ret, err
}

func (m *CNFServiceHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	svc := openwrt.SvcClient{OpenwrtClient: openwrtClient}
	service := instance.(*openwrt.SdewanSvc)
	return svc.CreateSvc(*service)
}

func (m *CNFServiceHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	svc := openwrt.SvcClient{OpenwrtClient: openwrtClient}
	service := instance.(*openwrt.SdewanSvc)
	return svc.UpdateSvc(*service)
}

func (m *CNFServiceHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	svc := openwrt.SvcClient{OpenwrtClient: openwrtClient}
	return svc.DeleteSvc(name)
}

func (m *CNFServiceHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

// CNFServiceReconciler reconciles a CNFService object
type CNFServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfservices/status,verbs=get;update;patch

func (r *CNFServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, cnfServiceHandler)
}

func (r *CNFServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFService{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.CNFServiceList{})),
			},
			Filter).
		Watches(
			&source.Kind{Type: &corev1.Service{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetServiceToRequestsFunc(r)),
			},
			IPFilter).
		Complete(r)
}
