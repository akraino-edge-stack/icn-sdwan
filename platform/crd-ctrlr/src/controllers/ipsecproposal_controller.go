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
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var ipsecProposalHandler = new(IpsecProposalHandler)

type IpsecProposalHandler struct {
}

func (m *IpsecProposalHandler) GetType() string {
	return "IpsecProposal"
}

func (m *IpsecProposalHandler) GetName(instance runtime.Object) string {
	proposal := instance.(*batchv1alpha1.IpsecProposal)
	return proposal.Name
}

func (m *IpsecProposalHandler) GetFinalizer() string {
	return "proposal.finalizers.sdewan.akraino.org"
}

func (m *IpsecProposalHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.IpsecProposal{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *IpsecProposalHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	proposal := instance.(*batchv1alpha1.IpsecProposal)
	proposal.Spec.Name = proposal.ObjectMeta.Name
	proposalObject := openwrt.SdewanIpsecProposal(proposal.Spec)
	return &proposalObject, nil
}

func (m *IpsecProposalHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	proposal1 := instance1.(*openwrt.SdewanIpsecProposal)
	proposal2 := instance2.(*openwrt.SdewanIpsecProposal)
	return reflect.DeepEqual(*proposal1, *proposal2)
}

func (m *IpsecProposalHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	ret, err := ipsec.GetProposal(name)
	return ret, err
}

func (m *IpsecProposalHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	proposal := instance.(*openwrt.SdewanIpsecProposal)
	return ipsec.CreateProposal(*proposal)
}

func (m *IpsecProposalHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	proposal := instance.(*openwrt.SdewanIpsecProposal)
	return ipsec.UpdateProposal(*proposal)
}

func (m *IpsecProposalHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	ipsec := openwrt.IpsecClient{OpenwrtClient: openwrtClient}
	return ipsec.DeleteProposal(name)
}

func (m *IpsecProposalHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("ipsec", "restart")
}

// IpsecProposalReconciler reconciles a IpsecProposal object
type IpsecProposalReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecproposals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=ipsecproposals/status,verbs=get;update;patch

func (r *IpsecProposalReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, ipsecProposalHandler)
}

func (r *IpsecProposalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.IpsecProposal{}, ps).
		Complete(r)
}
