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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mwan3conflog = logf.Log.WithName("mwan3conf-resource")

func (r *Mwan3Conf) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-batch-sdewan-akraino-org-v1alpha1-mwan3conf,mutating=true,failurePolicy=fail,groups=batch.sdewan.akraino.org,resources=mwan3confs,verbs=create;update,versions=v1alpha1,name=mmwan3conf.kb.io

var _ webhook.Defaulter = &Mwan3Conf{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Mwan3Conf) Default() {
	mwan3conflog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-batch-sdewan-akraino-org-v1alpha1-mwan3conf,mutating=false,failurePolicy=fail,groups=batch.sdewan.akraino.org,resources=mwan3confs,versions=v1alpha1,name=vmwan3conf.kb.io

var _ webhook.Validator = &Mwan3Conf{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Mwan3Conf) ValidateCreate() error {
	mwan3conflog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.ValidateMwan3()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Mwan3Conf) ValidateUpdate(old runtime.Object) error {
	mwan3conflog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.ValidateMwan3()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Mwan3Conf) ValidateDelete() error {
	mwan3conflog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Mwan3Conf) ValidateMwan3() error {
	policies := r.Spec.Policies
	for _, rOptions := range r.Spec.Rules {
		if _, ok := policies[rOptions.UsePolicy]; !ok {
			return &UnmatchError{"policy", rOptions.UsePolicy}
		}
	}
	return nil
}

type UnmatchError struct {
	RType string
	Name  string
}

func (err *UnmatchError) Error() string {
	return err.RType + " :" + err.Name + " doesnt' exist"
}
