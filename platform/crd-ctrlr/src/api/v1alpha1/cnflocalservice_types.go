// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFLocalServiceStatus defines the observed state of CNFLocalServiceStatus
type CNFLocalServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	LocalIP string `json:"localip,omitempty"`
	// +optional
	LocalPort string `json:"localport,omitempty"`
	// +optional
	RemoteIPs []string `json:"remoteips,omitempty"`
	// +optional
	RemotePort string `json:"remoteport,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

func (c *CNFLocalServiceStatus) IsEqual(s *CNFLocalServiceStatus) bool {
	if c.LocalIP != s.LocalIP ||
		c.LocalPort != s.LocalPort ||
		c.RemotePort != s.RemotePort {
		return false
	}
	if len(c.RemoteIPs) != len(s.RemoteIPs) {
		return false
	}

	for i := 0; i < len(c.RemoteIPs); i++ {
		if c.RemoteIPs[i] != s.RemoteIPs[i] {
			return false
		}
	}

	return true
}

// CNFLocalServiceSpec defines the desired state of CNFService
type CNFLocalServiceSpec struct {
	LocalService  string `json:"localservice,omitempty"`
	LocalPort     string `json:"localport,omitempty"`
	RemoteService string `json:"remoteservice,omitempty"`
	RemotePort    string `json:"remoteport,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFLocalService is the Schema for the cnflocalservices API
type CNFLocalService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFLocalServiceSpec   `json:"spec,omitempty"`
	Status CNFLocalServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFLocalServiceList contains a list of CNFLocalServiceList
type CNFLocalServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFLocalService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFLocalService{}, &CNFLocalServiceList{})
}
