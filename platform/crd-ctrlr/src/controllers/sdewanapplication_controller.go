// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package controllers

import (
	"context"
	"fmt"
	"log"
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

func (e *AppCRError) Error() string {
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
		err = r.List(ctx, podList, client.MatchingLabels(ps), client.InNamespace(ns))
		if err != nil {
			log.Println(err)
		}
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

		return podPhase == "Running"
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		podOldPhase := reflect.ValueOf(e.ObjectOld).Interface().(*corev1.Pod).Status.Phase
		podNewPhase := reflect.ValueOf(e.ObjectNew).Interface().(*corev1.Pod).Status.Phase

		if podOldPhase != podNewPhase && podNewPhase == "Running" {
			return true
		}

		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},
})

func GetAppToRequestsFunc(r client.Client) func(h handler.MapObject) []reconcile.Request {

	return func(h handler.MapObject) []reconcile.Request {
		podLabels := h.Meta.GetLabels()
		podNamespace := h.Meta.GetNamespace()
		appCRList := &batchv1alpha1.SdewanApplicationList{}
		cr := &batchv1alpha1.SdewanApplication{}
		ctx := context.Background()
		err := r.List(ctx, appCRList)
		if err != nil {
			log.Println(err)
		}
		crIsFound := false
		for _, appCR := range appCRList.Items {
			ps := appCR.Spec.PodSelector.MatchLabels
			ns := appCR.Spec.AppNamespace
			judge := true
			if ns != podNamespace {
				judge = false
			} else {
				for key, value := range ps {
					if _, ok := podLabels[key]; ok && podLabels[key] == value {
						continue
					} else {
						judge = false
						break
					}
				}
			}
			if judge {
				// Assume one application only have one Application CR
				cr = &appCR
				crIsFound = true
				break
			}
		}

		if crIsFound {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      cr.ObjectMeta.GetName(),
					Namespace: cr.ObjectMeta.GetNamespace(),
				}},
			}
		} else {
			return []reconcile.Request{}
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
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.SdewanApplicationList{})),
			},
			Filter).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetAppToRequestsFunc(r)),
			},
			appFilter).
		Complete(r)
}
