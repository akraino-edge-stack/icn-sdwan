// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFServiceSpec defines the desired state of CNFService
type CNFServiceSpec struct {
	FullName string `json:"fullname,omitempty"`
	Port     string `json:"port,omitempty"`
	DPort    string `json:"dport,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFService is the Schema for the cnfservices API
type CNFService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFServiceSpec `json:"spec,omitempty"`
	Status SdewanStatus   `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFServiceList contains a list of CNFService
type CNFServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFService{}, &CNFServiceList{})
}
