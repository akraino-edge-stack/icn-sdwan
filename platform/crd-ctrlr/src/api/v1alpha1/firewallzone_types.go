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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FirewallZoneSpec defines the desired state of FirewallZone
type FirewallZoneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of FirewallZone. Edit FirewallZone_types.go to remove/update
	Name             string   `json:"name,omitempty"`
	Network          []string `json:"network"`
	Masq             string   `json:"masq,omitempty"`
	MasqSrc          []string `json:"masq_src,omitempty"`
	MasqDest         []string `json:"masq_dest,omitempty"`
	MasqAllowInvalid string   `json:"masq_allow_invalid,omitempty"`
	MtuFix           string   `json:"mtu_fix,omitempty"`
	Input            string   `json:"input,omitempty"`
	Forward          string   `json:"forward,omitempty"`
	Output           string   `json:"output,omitempty"`
	Family           string   `json:"family,omitempty"`
	Subnet           []string `json:"subnet,omitempty"`
	ExtraSrc         string   `json:"extra_src,omitempty"`
	ExtraDest        string   `json:"etra_dest,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// FirewallZone is the Schema for the firewallzones API
type FirewallZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallZoneSpec `json:"spec,omitempty"`
	Status SdewanStatus     `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallZoneList contains a list of FirewallZone
type FirewallZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallZone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallZone{}, &FirewallZoneList{})
}
