// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFStatusSpec defines the desired state of CNFStatus
type CNFStatusSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// CNFStatusInformation defines the runtime information of a CNF
type CNFStatusInformation struct {
	Name      string `json:"name"`
	NameSpace string `json:"namespace,omitempty"`
	Node      string `json:"node,omitempty"`
	Purpose   string `json:"purpose,omitempty"`
	IP        string `json:"ip,omitempty"`
	Status    string `json:"status,omitempty"`
}

// CNFStatusStatus defines the observed state of CNFStatus
type CNFStatusStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	AppliedGeneration int64 `json:"appliedGeneration,omitempty"`
	// +optional
	AppliedTime *metav1.Time `json:"appliedTime,omitempty"`
	// +optional
	Information []CNFStatusInformation `json:"information,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFStatus is the Schema for the cnfstatuses API
type CNFStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFStatusSpec   `json:"spec,omitempty"`
	Status CNFStatusStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFStatusList contains a list of CNFStatus
type CNFStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFStatus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFStatus{}, &CNFStatusList{})
}
