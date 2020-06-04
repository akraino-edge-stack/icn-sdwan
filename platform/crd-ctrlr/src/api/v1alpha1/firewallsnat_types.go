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

// FirewallSNATSpec defines the desired state of FirewallSNAT
type FirewallSNATSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name     string `json:"name,omitempty"`
	Src      string `json:"src,omitempty"`
	SrcIp    string `json:"src_ip,omitempty"`
	SrcDIp   string `json:"src_dip,omitempty"`
	SrcMac   string `json:"src_mac,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	SrcDPort string `json:"src_dport,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Dest     string `json:"dest,omitempty"`
	DestIp   string `json:"dest_ip,omitempty"`
	DestPort string `json:"dest_port,omitempty"`
	Mark     string `json:"mark,omitempty"`
	Target   string `json:"target,omitempty"`
	Family   string `json:"family,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FirewallSNAT is the Schema for the firewallsnats API
type FirewallSNAT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallSNATSpec `json:"spec,omitempty"`
	Status SdewanStatus     `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallSNATList contains a list of FirewallSNAT
type FirewallSNATList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallSNAT `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallSNAT{}, &FirewallSNATList{})
}
