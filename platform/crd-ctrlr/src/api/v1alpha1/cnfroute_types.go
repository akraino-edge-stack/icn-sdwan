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

// CNFRouteSpec defines the desired state of CNFRoute
type CNFRouteSpec struct {
	Dst   string `json:"dst,omitempty"`
	Gw    string `json:"gw,omitempty"`
	Dev   string `json:"dev,omitempty"`
	Table string `json:"table,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFRoute is the Schema for the cnfroutes API
type CNFRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFRouteSpec `json:"spec,omitempty"`
	Status SdewanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFRouteList contains a list of CNFRoute
type CNFRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFRoute{}, &CNFRouteList{})
}
