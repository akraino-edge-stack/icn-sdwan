// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Mwan3PolicySpec defines the desired state of Mwan3Policy
type Mwan3PolicyMember struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Network string `json:"network"`
	Metric  int    `json:"metric"`
	Weight  int    `json:"weight"`
}

type Mwan3PolicySpec struct {
	Members []Mwan3PolicyMember `json:"members"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Mwan3Policy is the Schema for the mwan3policies API
type Mwan3Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Mwan3PolicySpec `json:"spec,omitempty"`
	Status SdewanStatus    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// Mwan3PolicyList contains a list of Mwan3Policy
type Mwan3PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mwan3Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mwan3Policy{}, &Mwan3PolicyList{})
}
