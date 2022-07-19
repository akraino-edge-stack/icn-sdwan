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

var networkFirewallRuleHandler = new(NetworkFirewallRuleHandler)

type NetworkFirewallRuleHandler struct {
}

func (m *NetworkFirewallRuleHandler) GetType() string {
	return "NetworkFirewallRule"
}

func (m *NetworkFirewallRuleHandler) GetName(instance client.Object) string {
	rule := instance.(*batchv1alpha1.NetworkFirewallRule)
	return rule.Name
}

func (m *NetworkFirewallRuleHandler) GetFinalizer() string {
	return "networkfirewallrule.finalizers.sdewan.akraino.org"
}

func (m *NetworkFirewallRuleHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (client.Object, error) {
	instance := &batchv1alpha1.NetworkFirewallRule{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *NetworkFirewallRuleHandler) Convert(instance client.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewallrule := instance.(*batchv1alpha1.NetworkFirewallRule)
	firewallrule.Spec.Name = firewallrule.ObjectMeta.Name
	firewallruleObject := openwrt.SdewanNetworkFirewallRule(firewallrule.Spec)
	return &firewallruleObject, nil
}

func (m *NetworkFirewallRuleHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	rule1 := instance1.(*openwrt.SdewanNetworkFirewallRule)
	rule2 := instance2.(*openwrt.SdewanNetworkFirewallRule)
	return reflect.DeepEqual(*rule1, *rule2)
}

func (m *NetworkFirewallRuleHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.NetworkFirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetRule(name)
	return ret, err
}

func (m *NetworkFirewallRuleHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.NetworkFirewallClient{OpenwrtClient: openwrtClient}
	rule := instance.(*openwrt.SdewanNetworkFirewallRule)
	return firewall.CreateRule(*rule)
}

func (m *NetworkFirewallRuleHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.NetworkFirewallClient{OpenwrtClient: openwrtClient}
	rule := instance.(*openwrt.SdewanNetworkFirewallRule)
	return firewall.UpdateRule(*rule)
}

func (m *NetworkFirewallRuleHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.NetworkFirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteRule(name)
}

func (m *NetworkFirewallRuleHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

// NetworkFirewallRuleReconciler reconciles a NetworkFirewallRule object
type NetworkFirewallRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=networkfirewallrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=networkfirewallrules/status,verbs=get;update;patch

func (r *NetworkFirewallRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r.Client, r.Log, ctx, req, networkFirewallRuleHandler)
}

func (r *NetworkFirewallRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.NetworkFirewallRule{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(GetToRequestsFunc(r.Client, &batchv1alpha1.NetworkFirewallRuleList{})),
			Filter).
		Complete(r)
}
