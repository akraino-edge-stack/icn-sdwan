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

var firewallZoneHandler = new(FirewallZoneHandler)

type FirewallZoneHandler struct {
}

func (m *FirewallZoneHandler) GetType() string {
	return "FirewallZone"
}

func (m *FirewallZoneHandler) GetName(instance client.Object) string {
	zone := instance.(*batchv1alpha1.FirewallZone)
	return zone.Name
}

func (m *FirewallZoneHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *FirewallZoneHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (client.Object, error) {
	instance := &batchv1alpha1.FirewallZone{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *FirewallZoneHandler) Convert(instance client.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewallzone := instance.(*batchv1alpha1.FirewallZone)
	instance_to_convert := batchv1alpha1.FirewallZoneSpec(firewallzone.Spec)
	networks := make([]string, len(instance_to_convert.Network))
	for index, network := range instance_to_convert.Network {
		if iface, err := net2iface(network, deployment); err != nil {
			return nil, err
		} else {
			networks[index] = iface
		}
	}
	instance_to_convert.Name = firewallzone.ObjectMeta.Name
	instance_to_convert.Network = networks
	firewallzoneObject := openwrt.SdewanFirewallZone(instance_to_convert)
	return &firewallzoneObject, nil
}

func (m *FirewallZoneHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	zone1 := instance1.(*openwrt.SdewanFirewallZone)
	zone2 := instance2.(*openwrt.SdewanFirewallZone)
	return reflect.DeepEqual(*zone1, *zone2)
}

func (m *FirewallZoneHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetZone(name)
	return ret, err
}

func (m *FirewallZoneHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	zone := instance.(*openwrt.SdewanFirewallZone)
	return firewall.CreateZone(*zone)
}

func (m *FirewallZoneHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	zone := instance.(*openwrt.SdewanFirewallZone)
	return firewall.UpdateZone(*zone)
}

func (m *FirewallZoneHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteZone(name)
}

func (m *FirewallZoneHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("firewall", "restart")
}

// FirewallZoneReconciler reconciles a FirewallZone object
type FirewallZoneReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallzones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallzones/status,verbs=get;update;patch

func (r *FirewallZoneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r.Client, r.Log, ctx, req, firewallZoneHandler)
}

func (r *FirewallZoneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.FirewallZone{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(GetToRequestsFunc(r.Client, &batchv1alpha1.FirewallZoneList{})),
			Filter).
		Complete(r)
}
