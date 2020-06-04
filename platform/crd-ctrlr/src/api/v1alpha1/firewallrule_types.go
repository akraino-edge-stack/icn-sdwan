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

// FirewallRuleSpec defines the desired state of FirewallRule
type FirewallRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of FirewallRule. Edit FirewallRule_types.go to remove/update
	Name     string   `json:"name,omitempty"`
	Src      string   `json:"src,omitempty"`
	SrcIp    string   `json:"src_ip,omitempty"`
	SrcMac   string   `json:"src_mac,omitempty"`
	SrcPort  string   `json:"src_port,omitempty"`
	Proto    string   `json:"proto,omitempty"`
	IcmpType []string `json:"icmp_type,omitempty"`
	Dest     string   `json:"dest,omitempty"`
	DestIp   string   `json:"dest_ip,omitempty"`
	DestPort string   `json:"dest_port,omitempty"`
	Mark     string   `json:"mark,omitempty"`
	Target   string   `json:"target,omitempty"`
	SetMark  string   `json:"set_mark,omitempty"`
	SetXmark string   `json:"set_xmark,omitempty"`
	Family   string   `json:"family,omitempty"`
	Extra    string   `json:"extra,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FirewallRule is the Schema for the firewallrules API
type FirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallRuleSpec `json:"spec,omitempty"`
	Status SdewanStatus     `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallRuleList contains a list of FirewallRule
type FirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallRule{}, &FirewallRuleList{})
}
