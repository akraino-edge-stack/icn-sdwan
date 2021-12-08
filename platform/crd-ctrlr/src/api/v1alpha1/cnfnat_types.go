// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFNATSpec defines the desired state of CNFNAT
type CNFNATSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name     string `json:"name,omitempty"`
	Src      string `json:"src,omitempty"`
	SrcIp    string `json:"src_ip,omitempty"`
	SrcDIp   string `json:"src_dip,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	SrcDPort string `json:"src_dport,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Dest     string `json:"dest,omitempty"`
	DestIp   string `json:"dest_ip,omitempty"`
	DestPort string `json:"dest_port,omitempty"`
	Target   string `json:"target,omitempty"`
	Index    string `json:"index,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFNAT is the Schema for the cnfnats API
type CNFNAT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFNATSpec   `json:"spec,omitempty"`
	Status SdewanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFNATList contains a list of CNFNAT
type CNFNATList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFNAT `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFNAT{}, &CNFNATList{})
}
