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

var firewallSnatHandler = new(FirewallSnatHandler)

type FirewallSnatHandler struct {
}

func (m *FirewallSnatHandler) GetType() string {
	return "FirewallSnat"
}

func (m *FirewallSnatHandler) GetName(instance runtime.Object) string {
	snat := instance.(*batchv1alpha1.FirewallSNAT)
	return snat.Name
}

func (m *FirewallSnatHandler) GetFinalizer() string {
	return "snat.finalizers.sdewan.akraino.org"
}

func (m *FirewallSnatHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.FirewallSNAT{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

//pupulate "snat" to target field as default value
//copy "name" field value from metadata to SPEC.name
func (m *FirewallSnatHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewallsnat := instance.(*batchv1alpha1.FirewallSNAT)
	firewallsnat.Spec.Name = firewallsnat.ObjectMeta.Name
	firewallsnat.Spec.Target = "SNAT"
	firewallsnatObject := openwrt.SdewanFirewallRedirect(firewallsnat.Spec)
	return &firewallsnatObject, nil
}

func (m *FirewallSnatHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	snat1 := instance1.(*openwrt.SdewanFirewallRedirect)
	snat2 := instance2.(*openwrt.SdewanFirewallRedirect)
	return reflect.DeepEqual(*snat1, *snat2)
}

func (m *FirewallSnatHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetRedirect(name)
	return ret, err
}

func (m *FirewallSnatHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	snat := instance.(*openwrt.SdewanFirewallRedirect)
	return firewall.CreateRedirect(*snat)
}

func (m *FirewallSnatHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	snat := instance.(*openwrt.SdewanFirewallRedirect)
	return firewall.UpdateRedirect(*snat)
}

func (m *FirewallSnatHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteRedirect(name)
}

func (m *FirewallSnatHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("firewall", "restart")
}

// FirewallSNATReconciler reconciles a FirewallSNAT object
type FirewallSNATReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallsnats,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallsnats/status,verbs=get;update;patch

func (r *FirewallSNATReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, firewallSnatHandler)
}

func (r *FirewallSNATReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.FirewallSNAT{}, ps).
		Complete(r)
}
