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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var firewallForwardingHandler = new(FirewallForwardingHandler)

type FirewallForwardingHandler struct {
}

func (m *FirewallForwardingHandler) GetType() string {
	return "FirewallForwarding"
}

func (m *FirewallForwardingHandler) GetName(instance runtime.Object) string {
	forwarding := instance.(*batchv1alpha1.FirewallForwarding)
	return forwarding.Name
}

func (m *FirewallForwardingHandler) GetFinalizer() string {
	return "forwarding.finalizers.sdewan.akraino.org"
}

func (m *FirewallForwardingHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.FirewallForwarding{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *FirewallForwardingHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewallforwarding := instance.(*batchv1alpha1.FirewallForwarding)
	firewallforwarding.Spec.Name = firewallforwarding.ObjectMeta.Name
	firewallforwardingObject := openwrt.SdewanFirewallForwarding(firewallforwarding.Spec)
	return &firewallforwardingObject, nil
}

func (m *FirewallForwardingHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	forwarding1 := instance1.(*openwrt.SdewanFirewallForwarding)
	forwarding2 := instance2.(*openwrt.SdewanFirewallForwarding)
	return reflect.DeepEqual(*forwarding1, *forwarding2)
}

func (m *FirewallForwardingHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetForwarding(name)
	return ret, err
}

func (m *FirewallForwardingHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	forwarding := instance.(*openwrt.SdewanFirewallForwarding)
	return firewall.CreateForwarding(*forwarding)
}

func (m *FirewallForwardingHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	forwarding := instance.(*openwrt.SdewanFirewallForwarding)
	return firewall.UpdateForwarding(*forwarding)
}

func (m *FirewallForwardingHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteForwarding(name)
}

func (m *FirewallForwardingHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("firewall", "restart")
}

// FirewallForwardingReconciler reconciles a FirewallForwarding object
type FirewallForwardingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallforwardings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallforwardings/status,verbs=get;update;patch

func (r *FirewallForwardingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, firewallForwardingHandler)
}

func (r *FirewallForwardingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.FirewallForwarding{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.FirewallForwardingList{})),
			},
			Filter).
		Complete(r)
}
