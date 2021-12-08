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

var cnfnatHandler = new(CNFNatHandler)

type CNFNatHandler struct {
}

func (m *CNFNatHandler) GetType() string {
	return "CNFNAT"
}

func (m *CNFNatHandler) GetName(instance runtime.Object) string {
	nat := instance.(*batchv1alpha1.CNFNAT)
	return nat.Name
}

func (m *CNFNatHandler) GetFinalizer() string {
	return "cnfnat.finalizers.sdewan.akraino.org"
}

func (m *CNFNatHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.CNFNAT{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

//pupulate "nat" to target field as default value
func (m *CNFNatHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	cnfnat := instance.(*batchv1alpha1.CNFNAT)
	cnfnat.Spec.Name = cnfnat.ObjectMeta.Name
	cnfnatObject := openwrt.SdewanNat(cnfnat.Spec)
	return &cnfnatObject, nil
}

func (m *CNFNatHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	nat1 := instance1.(*openwrt.SdewanNat)
	nat2 := instance2.(*openwrt.SdewanNat)
	return reflect.DeepEqual(*nat1, *nat2)
}

func (m *CNFNatHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	natClient := openwrt.NatClient{OpenwrtClient: openwrtClient}
	ret, err := natClient.GetNat(name)
	return ret, err
}

func (m *CNFNatHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	natClient := openwrt.NatClient{OpenwrtClient: openwrtClient}
	nat := instance.(*openwrt.SdewanNat)
	return natClient.CreateNat(*nat)
}

func (m *CNFNatHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	natClient := openwrt.NatClient{OpenwrtClient: openwrtClient}
	nat := instance.(*openwrt.SdewanNat)
	return natClient.UpdateNat(*nat)
}

func (m *CNFNatHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	natClient := openwrt.NatClient{OpenwrtClient: openwrtClient}
	return natClient.DeleteNat(name)
}

func (m *CNFNatHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

// CNFNATReconciler reconciles a CNFNAT object
type CNFNATReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfnats,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfnats/status,verbs=get;update;patch

func (r *CNFNATReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, cnfnatHandler)
}

func (r *CNFNATReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFNAT{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.CNFNATList{})),
			},
			Filter).
		Complete(r)
}
