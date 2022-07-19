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

var ipsecHostHandler = new(IpsecHostHandler)

type IpsecHostHandler struct {
}

func (m *IpsecHostHandler) GetType() string {
	return "IpsecHost"
}

func (m *IpsecHostHandler) GetName(instance client.Object) string {
	host := instance.(*batchv1alpha1.IpsecHost)
	return host.Name
}

func (m *IpsecHostHandler) GetFinalizer() string {
	return "ipsec.host.finalizers.sdewan.akraino.org"
}

func (m *IpsecHostHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (client.Object, error) {
	instance := &batchv1alpha1.IpsecHost{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *IpsecHostHandler) Convert(instance client.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	host := instance.(*batchv1alpha1.IpsecHost)
	numOfConn := len(host.Spec.Connections)
	conn := host.Spec.Connections
	openwrtConn := make([]openwrt.SdewanIpsecConnection, numOfConn)
	for i := 0; i < numOfConn; i++ {
		openwrtConn[i] = openwrt.SdewanIpsecConnection{
			Name:           conn[i].Name,
			ConnType:       conn[i].ConnectionType,
			Mode:           conn[i].Mode,
			LocalSourceip:  conn[i].LocalSourceIp,
			LocalUpdown:    conn[i].LocalUpDown,
			LocalFirewall:  conn[i].LocalFirewall,
			RemoteSubnet:   conn[i].RemoteSubnet,
			RemoteSourceip: conn[i].RemoteSourceIp,
			RemoteUpdown:   conn[i].RemoteUpDown,
			RemoteFirewall: conn[i].RemoteFirewall,
			CryptoProposal: conn[i].CryptoProposal,
			Mark:           conn[i].Mark,
			IfId:           conn[i].IfId,
		}
	}
	hostObject := openwrt.SdewanIpsecRemote{
		Name:                 host.Name,
		Gateway:              host.Spec.Remote,
		Type:                 host.Spec.Type,
		AuthenticationMethod: host.Spec.AuthenticationMethod,
		PreSharedKey:         host.Spec.PresharedKey,
		LocalIdentifier:      host.Spec.LocalIdentifier,
		RemoteIdentifier:     host.Spec.RemoteIdentifier,
		CryptoProposal:       host.Spec.CryptoProposal,
		ForceCryptoProposal:  host.Spec.ForceCryptoProposal,
		LocalPublicCert:      host.Spec.LocalPublicCert,
		LocalPrivateCert:     host.Spec.LocalPrivateCert,
		SharedCa:             host.Spec.SharedCA,
		Connections:          openwrtConn,
	}
	return &hostObject, nil
}

func (m *IpsecHostHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	host1 := instance1.(*openwrt.SdewanIpsecRemote)
	host2 := instance2.(*openwrt.SdewanIpsecRemote)
	return reflect.DeepEqual(*host1, *host2)
}

func (m *IpsecHostHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	ret, err := ipsec.GetRemote(name)
	return ret, err
}

func (m *IpsecHostHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	host := instance.(*openwrt.SdewanIpsecRemote)
	return ipsec.CreateRemote(*host)
}

func (m *IpsecHostHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	host := instance.(*openwrt.SdewanIpsecRemote)
	return ipsec.UpdateRemote(*host)
}

func (m *IpsecHostHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	return ipsec.DeleteRemote(name)
}

func (m *IpsecHostHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("ipsec", "restart")
}

// IpsecHostReconciler reconciles a IpsecHost object
type IpsecHostReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsechosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsechosts/status,verbs=get;update;patch

func (r *IpsecHostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r.Client, r.Log, ctx, req, ipsecHostHandler)
}

func (r *IpsecHostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.IpsecHost{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(GetToRequestsFunc(r.Client, &batchv1alpha1.IpsecHostList{})),
			Filter).
		Complete(r)
}
