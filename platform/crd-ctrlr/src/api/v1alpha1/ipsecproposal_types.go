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

// IpsecProposalSpec defines the desired state of IpsecProposal
type IpsecProposalSpec struct {
	Name                string `json:"name,omitempty"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	HashAlgorithm       string `json:"hash_algorithm"`
	DhGroup             string `json:"dh_group"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IpsecProposal is the Schema for the ipsecproposals API
type IpsecProposal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpsecProposalSpec `json:"spec,omitempty"`
	Status SdewanStatus      `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IpsecProposalList contains a list of IpsecProposal
type IpsecProposalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpsecProposal `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpsecProposal{}, &IpsecProposalList{})
}
