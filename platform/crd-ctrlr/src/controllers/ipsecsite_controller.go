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

var ipsecSiteHandler = new(IpsecSiteHandler)

type IpsecSiteHandler struct {
}

func (m *IpsecSiteHandler) GetType() string {
	return "IpsecSite"
}

func (m *IpsecSiteHandler) GetName(instance runtime.Object) string {
	site := instance.(*batchv1alpha1.IpsecSite)
	return site.Name
}

func (m *IpsecSiteHandler) GetFinalizer() string {
	return "ipsec.site.finalizers.sdewan.akraino.org"
}

func (m *IpsecSiteHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.IpsecSite{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *IpsecSiteHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	site := instance.(*batchv1alpha1.IpsecSite)
	numOfConn := len(site.Spec.Connections)
	conn := site.Spec.Connections
	openwrtConn := make([]openwrt.SdewanIpsecConnection, numOfConn)
	for i := 0; i < numOfConn; i++ {
		openwrtConn[i] = openwrt.SdewanIpsecConnection{
			Name:           conn[i].Name,
			ConnType:       conn[i].ConnectionType,
			Mode:           conn[i].Mode,
			LocalSubnet:    conn[i].LocalSubnet,
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
	siteObject := openwrt.SdewanIpsecRemote{
		Name:                 site.Name,
		Gateway:              site.Spec.Remote,
		Type:                 site.Spec.Type,
		AuthenticationMethod: site.Spec.AuthenticationMethod,
		PreSharedKey:         site.Spec.PresharedKey,
		LocalIdentifier:      site.Spec.LocalIdentifier,
		RemoteIdentifier:     site.Spec.RemoteIdentifier,
		CryptoProposal:       site.Spec.CryptoProposal,
		ForceCryptoProposal:  site.Spec.ForceCryptoProposal,
		LocalPublicCert:      site.Spec.LocalPublicCert,
		LocalPrivateCert:     site.Spec.LocalPrivateCert,
		SharedCa:             site.Spec.SharedCA,
		Connections:          openwrtConn,
	}
	return &siteObject, nil
}

func (m *IpsecSiteHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	site1 := instance1.(*openwrt.SdewanIpsecRemote)
	site2 := instance2.(*openwrt.SdewanIpsecRemote)
	return reflect.DeepEqual(*site1, *site2)
}

func (m *IpsecSiteHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	ret, err := ipsec.GetRemote(name)
	return ret, err
}

func (m *IpsecSiteHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	site := instance.(*openwrt.SdewanIpsecRemote)
	return ipsec.CreateRemote(*site)
}

func (m *IpsecSiteHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	site := instance.(*openwrt.SdewanIpsecRemote)
	return ipsec.UpdateRemote(*site)
}

func (m *IpsecSiteHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	return ipsec.DeleteRemote(name)
}

func (m *IpsecSiteHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("ipsec", "restart")
}

// IpsecSiteReconciler reconciles a IpsecSite object
type IpsecSiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecsites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecsites/status,verbs=get;update;patch

func (r *IpsecSiteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, ipsecSiteHandler)
}

func (r *IpsecSiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.IpsecSite{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.IpsecSiteList{})),
			},
			Filter).
		Complete(r)
}
