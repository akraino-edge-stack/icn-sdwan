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

type Network struct {
	Name string `json:"name"`

	// +optional
	IsProvider bool `json:"isProvider"`

	// +optional
	Interface string `json:"interface"`

	// +optional
	DefaultGateway bool `json:"defaultGateway"`
}

// SdewanSpec defines the desired state of Sdewan
type SdewanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Sdewan. Edit Sdewan_types.go to remove/update
	Node     string    `json:"node"`
	Networks []Network `json:"networks"`
	// +optional
	Mwan3Conf string `json:"mwan3Conf"`
}

type Mwan3Status struct {
	Name string `json:"name"`

	IsApplied bool `json:"isApplied"`

	// +optional
	// +nullable
	AppliedTime *metav1.Time `json:"appliedTime"`
}

// SdewanStatus defines the observed state of Sdewan
type SdewanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Mwan3Status Mwan3Status `json:"mwan3Status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Sdewan is the Schema for the sdewans API
type Sdewan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SdewanSpec   `json:"spec,omitempty"`
	Status SdewanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SdewanList contains a list of Sdewan
type SdewanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sdewan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sdewan{}, &SdewanList{})
}
