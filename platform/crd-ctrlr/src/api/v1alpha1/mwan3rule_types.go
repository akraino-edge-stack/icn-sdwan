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

// Mwan3RuleSpec defines the desired state of Mwan3Rule

type Mwan3RuleSpec struct {
	// Members []Mwan3Rule `json:"members"`
	Policy   string `json:"policy"`
	SrcIp    string `json:"src_ip"`
	SrcPort  string `json:"src_port"`
	DestIp   string `json:"dest_ip"`
	DestPort string `json:"dest_port"`
	Proto    string `json:"proto"`
	Family   string `json:"family"`
	Sticky   string `json:"sticky"`
	Timeout  string `json:"timeout"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Mwan3Rule is the Schema for the mwan3rules API
type Mwan3Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Mwan3RuleSpec `json:"spec,omitempty"`
	Status SdewanStatus  `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// Mwan3RuleList contains a list of Mwan3Rule
type Mwan3RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mwan3Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mwan3Rule{}, &Mwan3RuleList{})
}
