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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var mwan3RuleHandler = new(Mwan3RuleHandler)

type Mwan3RuleHandler struct {
}

func (m *Mwan3RuleHandler) GetType() string {
	return "Mwan3Rule"
}

func (m *Mwan3RuleHandler) GetName(instance runtime.Object) string {
	Rule := instance.(*batchv1alpha1.Mwan3Rule)
	return Rule.Name
}

func (m *Mwan3RuleHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *Mwan3RuleHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.Mwan3Rule{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *Mwan3RuleHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	rule := instance.(*batchv1alpha1.Mwan3Rule)
	// openwrtrule := openwrt.SdewanRule{
	// 	rule.Spec}
	openwrtrule := openwrt.SdewanRule{
		Name:     rule.Name,
		Policy:   rule.Spec.Policy,
		SrcIp:    rule.Spec.SrcIp,
		SrcPort:  rule.Spec.SrcPort,
		DestIp:   rule.Spec.DestIp,
		DestPort: rule.Spec.DestPort,
		Proto:    rule.Spec.Proto,
		Family:   rule.Spec.Family,
		Sticky:   rule.Spec.Sticky,
		Timeout:  rule.Spec.Timeout,
	}
	return &openwrtrule, nil
	// return &openwrt.SdewanRule(instance), nil
}

func (m *Mwan3RuleHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	Rule1 := instance1.(*openwrt.SdewanRule)
	Rule2 := instance2.(*openwrt.SdewanRule)
	return reflect.DeepEqual(*Rule1, *Rule2)
}

func (m *Mwan3RuleHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	ret, err := mwan3.GetRule(name)
	printjson(ret)
	return ret, err
}

func (m *Mwan3RuleHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	Rule := instance.(*openwrt.SdewanRule)
	return mwan3.CreateRule(*Rule)
}

func (m *Mwan3RuleHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	Rule := instance.(*openwrt.SdewanRule)
	return mwan3.UpdateRule(*Rule)
}

func (m *Mwan3RuleHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	return mwan3.DeleteRule(name)
}

func (m *Mwan3RuleHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	return service.ExecuteService("mwan3", "restart")
}

// Mwan3RuleReconciler reconciles a Mwan3Rule object
type Mwan3RuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=mwan3rules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=mwan3rules/status,verbs=get;update;patch

func (r *Mwan3RuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, mwan3RuleHandler)
}

func (r *Mwan3RuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.Mwan3Rule{}).
		Complete(r)
}
