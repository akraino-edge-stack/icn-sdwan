// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var cnfRouteRuleHandler = new(CNFRouteRuleHandler)

type CNFRouteRuleHandler struct {
}

func (m *CNFRouteRuleHandler) GetType() string {
	return "cnfRouteRule"
}

func (m *CNFRouteRuleHandler) GetName(instance runtime.Object) string {
	routerule := instance.(*batchv1alpha1.CNFRouteRule)
	return routerule.Name
}

func (m *CNFRouteRuleHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *CNFRouteRuleHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.CNFRouteRule{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *CNFRouteRuleHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	routerule := instance.(*batchv1alpha1.CNFRouteRule)
	openwrtrouterule := openwrt.SdewanRouteRule{
		Name:   routerule.Name,
		Src:    routerule.Spec.Src,
		Dst:    routerule.Spec.Dst,
		Flag:   routerule.Spec.Not,
		Prio:   routerule.Spec.Prio,
		Fwmark: routerule.Spec.Fwmark,
		Table:  routerule.Spec.Table,
	}
	return &openwrtrouterule, nil
}

func (m *CNFRouteRuleHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	routerule1 := instance1.(*openwrt.SdewanRouteRule)
	routerule2 := instance2.(*openwrt.SdewanRouteRule)
	return reflect.DeepEqual(*routerule1, *routerule2)
}

func (m *CNFRouteRuleHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	routerule := openwrt.RouteRuleClient{OpenwrtClient: openwrtClient}
	ret, err := routerule.GetRouteRule(name)
	return ret, err
}

func (m *CNFRouteRuleHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	routerule := openwrt.RouteRuleClient{OpenwrtClient: openwrtClient}
	obj := instance.(*openwrt.SdewanRouteRule)
	return routerule.CreateRouteRule(*obj)
}

func (m *CNFRouteRuleHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	routerule := openwrt.RouteRuleClient{OpenwrtClient: openwrtClient}
	obj := instance.(*openwrt.SdewanRouteRule)
	return routerule.UpdateRouteRule(*obj)
}

func (m *CNFRouteRuleHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	routerule := openwrt.RouteRuleClient{OpenwrtClient: openwrtClient}
	return routerule.DeleteRouteRule(name)
}

func (m *CNFRouteRuleHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

// CNFRouteRuleReconciler reconciles a CNFRouteRule object
type CNFRouteRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfrouterules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfrouterules/status,verbs=get;update;patch

func (r *CNFRouteRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, cnfRouteRuleHandler)
}

func (r *CNFRouteRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFRouteRule{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.CNFRouteRuleList{})),
			},
			Filter).
		Complete(r)
}
