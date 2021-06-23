// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFRouteRuleSpec defines the desired state of CNFRouteRule
type CNFRouteRuleSpec struct {
	// +kubebuilder:validation:Default:=""
	Src string `json:"src,omitempty"`
	// +kubebuilder:validation:Default:=""
	Dst string `json:"dst,omitempty"`
	// +kubebuilder:validation:Default:=false
	Not bool `json:"not,omitempty"`
	// +kubebuilder:validation:Default:=""
	Prio string `json:"prio,omitempty"`
	// +kubebuilder:validation:Default:=""
	Fwmark string `json:"fwmark,omitempty"`
	// +kubebuilder:validation:Default:=""
	Table string `json:"table,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFRouteRule is the Schema for the cnfrouterules API
type CNFRouteRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFRouteRuleSpec `json:"spec,omitempty"`
	Status SdewanStatus     `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFRouteRuleList contains a list of CNFRouteRule
type CNFRouteRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFRouteRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFRouteRule{}, &CNFRouteRuleList{})
}
