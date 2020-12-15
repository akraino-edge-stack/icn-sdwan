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
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var sdewanApplicationHandler = new(SdewanApplicationHandler)

type SdewanApplicationHandler struct {
}

type AppCRError struct {
	Code    int
	Message string
}

func (e AppCRError) Error() string {
	return fmt.Sprintf("Error Code: %d, Error Message: %s", e.Code, e.Message)
}

func (m *SdewanApplicationHandler) GetType() string {
	return "sdewanApplication"
}

func (m *SdewanApplicationHandler) GetName(instance runtime.Object) string {
	app := instance.(*batchv1alpha1.SdewanApplication)
	return app.Name
}

func (m *SdewanApplicationHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *SdewanApplicationHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.SdewanApplication{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err == nil {
		ps := instance.Spec.PodSelector.MatchLabels
		ns := instance.Spec.AppNamespace
		podList := &corev1.PodList{}
		r.List(ctx, podList, client.MatchingLabels(ps), client.InNamespace(ns))
		ips := ""
		for _, item := range podList.Items {
			if ips == "" {
				ips = item.Status.PodIP
			} else {
				ips = ips + "," + item.Status.PodIP
			}
		}
		instance.AppInfo.IpList = ips
	}

	if instance.AppInfo.IpList == "" {
		return instance, &AppCRError{Code: 404, Message: "Application not found"}
	}

	return instance, err
}

func (m *SdewanApplicationHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	app := instance.(*batchv1alpha1.SdewanApplication)
	openwrtapp := openwrt.SdewanApp{
		Name:   app.Name,
		IpList: app.AppInfo.IpList,
	}
	return &openwrtapp, nil
}

func (m *SdewanApplicationHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	app1 := instance1.(*openwrt.SdewanApp)
	app2 := instance2.(*openwrt.SdewanApp)
	return reflect.DeepEqual(*app1, *app2)
}

func (m *SdewanApplicationHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	app := openwrt.AppClient{OpenwrtClient: openwrtClient}
	ret, err := app.GetApp(name)
	return ret, err
}

func (m *SdewanApplicationHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	app := openwrt.AppClient{OpenwrtClient: openwrtClient}
	application := instance.(*openwrt.SdewanApp)
	return app.CreateApp(*application)
}

func (m *SdewanApplicationHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	app := openwrt.AppClient{OpenwrtClient: openwrtClient}
	application := instance.(*openwrt.SdewanApp)
	return app.UpdateApp(*application)
}

func (m *SdewanApplicationHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	app := openwrt.AppClient{OpenwrtClient: openwrtClient}
	return app.DeleteApp(name)
}

func (m *SdewanApplicationHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

var appFilter = builder.WithPredicates(predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		podPhase := reflect.ValueOf(e.Object).Interface().(*corev1.Pod).Status.Phase

		if podPhase == "Running" {
			return true
		}
		return false
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		podOldPhase := reflect.ValueOf(e.ObjectOld).Interface().(*corev1.Pod).Status.Phase
		podNewPhase := reflect.ValueOf(e.ObjectNew).Interface().(*corev1.Pod).Status.Phase

		if podOldPhase != podNewPhase && podNewPhase == "Running" {
			return true
		}
		// TODO
		return false
	},
})

func GetAppToRequestsFunc(r client.Client) func(h handler.MapObject) []reconcile.Request {

	return func(h handler.MapObject) []reconcile.Request {
		podLabels := h.Meta.GetLabels()
		appCRList := &batchv1alpha1.SdewanApplicationList{}
		cr := &batchv1alpha1.SdewanApplication{}
		ctx := context.Background()
		r.List(ctx, appCRList)
		for _, appCR := range appCRList.Items {
			ps := appCR.Spec.PodSelector.MatchLabels
			judge := true
			for key, value := range ps {
				if _, ok := podLabels[key]; ok && podLabels[key] == value {
					continue
				} else {
					judge = false
					break
				}
			}
			if judge {
				// Assume one application only have one Application CR
				cr = &appCR
				break
			}
		}

		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      cr.ObjectMeta.GetName(),
				Namespace: cr.ObjectMeta.GetNamespace(),
			}},
		}
	}
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
	return ProcessReconcile(r, r.Log, req, sdewanApplicationHandler)
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
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetAppToRequestsFunc(r)),
			},
			appFilter).
		Complete(r)
}
