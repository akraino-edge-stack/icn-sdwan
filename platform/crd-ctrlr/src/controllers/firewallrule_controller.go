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

var firewallRuleHandler = new(FirewallRuleHandler)

type FirewallRuleHandler struct {
}

func (m *FirewallRuleHandler) GetType() string {
	return "FirewallRule"
}

func (m *FirewallRuleHandler) GetName(instance runtime.Object) string {
	rule := instance.(*batchv1alpha1.FirewallRule)
	return rule.Name
}

func (m *FirewallRuleHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *FirewallRuleHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.FirewallRule{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *FirewallRuleHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	firewallrule := instance.(*batchv1alpha1.FirewallRule)
	firewallrule.Spec.Name = firewallrule.ObjectMeta.Name
	firewallruleObject := openwrt.SdewanFirewallRule(firewallrule.Spec)
	return &firewallruleObject, nil
}

func (m *FirewallRuleHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	rule1 := instance1.(*openwrt.SdewanFirewallRule)
	rule2 := instance2.(*openwrt.SdewanFirewallRule)
	return reflect.DeepEqual(*rule1, *rule2)
}

func (m *FirewallRuleHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	ret, err := firewall.GetRule(name)
	return ret, err
}

func (m *FirewallRuleHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	rule := instance.(*openwrt.SdewanFirewallRule)
	return firewall.CreateRule(*rule)
}

func (m *FirewallRuleHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	rule := instance.(*openwrt.SdewanFirewallRule)
	return firewall.UpdateRule(*rule)
}

func (m *FirewallRuleHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	firewall := openwrt.FirewallClient{OpenwrtClient: openwrtClient}
	return firewall.DeleteRule(name)
}

func (m *FirewallRuleHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("firewall", "restart")
}

// FirewallRuleReconciler reconciles a FirewallRule object
type FirewallRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=firewallrules/status,verbs=get;update;patch

func (r *FirewallRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, firewallRuleHandler)
}

func (r *FirewallRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.FirewallRule{}, ps).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.FirewallRuleList{})),
			},
			Filter).
		Complete(r)
}
