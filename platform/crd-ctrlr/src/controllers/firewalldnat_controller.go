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

var firewallDnatHandler = new(FirewallDnatHandler)

type FirewallDnatHandler struct {
}

func (m *FirewallDnatHandler) GetType() string {
	return "FirewallDnat"
}

func (m *FirewallDnatHandler) GetName(instance runtime.Object) string {
	dnat := instance.(*batchv1alpha1.FirewallDNAT)
	return dnat.Name
}

func (m *FirewallDnatHandler) GetFinalizer() string {
	return "dnat.finalizers.sdewan.akraino.org"
}

func (m *FirewallDnatHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.FirewallDNAT{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

//pupulate "dnat" to target field as default value
//copy "name" field value from metadata to SPEC.name
func (m *FirewallDnatHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewalldnat := instance.(*batchv1alpha1.FirewallDNAT)
	firewalldnat.Spec.Name = firewalldnat.ObjectMeta.Name
	firewalldnat.Spec.Target = "DNAT"
	firewalldnatObject := openwrt.SdewanFirewallRedirect(firewalldnat.Spec)
	return &firewalldnatObject, nil
}

func (m *FirewallDnatHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	dnat1 := instance1.(*openwrt.SdewanFirewallRedirect)
	dnat2 := instance2.(*openwrt.SdewanFirewallRedirect)
	return reflect.DeepEqual(*dnat1, *dnat2)
}

func (m *FirewallDnatHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetRedirect(name)
	return ret, err
}

func (m *FirewallDnatHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	dnat := instance.(*openwrt.SdewanFirewallRedirect)
	return firewall.CreateRedirect(*dnat)
}

func (m *FirewallDnatHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	dnat := instance.(*openwrt.SdewanFirewallRedirect)
	return firewall.UpdateRedirect(*dnat)
}

func (m *FirewallDnatHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteRedirect(name)
}

func (m *FirewallDnatHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("firewall", "restart")
}

// FirewallDNATReconciler reconciles a FirewallDNAT object
type FirewallDNATReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewalldnats,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewalldnats/status,verbs=get;update;patch

func (r *FirewallDNATReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, firewallDnatHandler)
}

func (r *FirewallDNATReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.FirewallDNAT{}, ps).
		Complete(r)
}
