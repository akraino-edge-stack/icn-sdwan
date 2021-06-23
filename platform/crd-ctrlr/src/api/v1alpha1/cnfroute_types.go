// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFRouteSpec defines the desired state of CNFRoute
type CNFRouteSpec struct {
	Dst string `json:"dst,omitempty"`
	Gw  string `json:"gw,omitempty"`
	Dev string `json:"dev,omitempty"`
	// +kubebuilder:validation:Enum=default;cnf
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
