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

type PolicyMember struct {
	Network string `json:"network"`
	Metric  int    `json:"metric"`
	Weight  int    `json:"weight"`
}

type PolicyMembers struct {
	Members []PolicyMember `json:"members"`
}
type Rule struct {
	UsePolicy string `json:"use_policy"`

	// +optional
	DestIP string `json:"dest_ip"`

	// +optional
	DestPort string `json:"dest_port"`

	// +optional
	SrcIP string `json:"src_IP"`

	// +optional
	SrcPort string `json:"src_port"`

	// +optional
	Proto string `json:"proto"`

	// +optional
	Family string `json:"family"`

	// +optional
	Sticky string `json:"sticky"`

	// +optional
	Timeout string `json:"timeout"`
}

// Mwan3ConfSpec defines the desired state of Mwan3Conf
type Mwan3ConfSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Policies map[string]PolicyMembers `json:"policy"`
	Rules    map[string]Rule          `json:"rule"`
}

// Mwan3ConfStatus defines the observed state of Mwan3Conf
type Mwan3ConfStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Mwan3Conf is the Schema for the mwan3confs API
type Mwan3Conf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Mwan3ConfSpec   `json:"spec,omitempty"`
	Status Mwan3ConfStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// Mwan3ConfList contains a list of Mwan3Conf
type Mwan3ConfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mwan3Conf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mwan3Conf{}, &Mwan3ConfList{})
}
