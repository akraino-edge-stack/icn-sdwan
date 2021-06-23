// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FirewallForwardingSpec defines the desired state of FirewallForwarding
type FirewallForwardingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name   string `json:"name,omitempty"`
	Src    string `json:"src,omitempty"`
	Dest   string `json:"dest,omitempty"`
	Family string `json:"family,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FirewallForwarding is the Schema for the firewallforwardings API
type FirewallForwarding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallForwardingSpec `json:"spec,omitempty"`
	Status SdewanStatus           `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallForwardingList contains a list of FirewallForwarding
type FirewallForwardingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallForwarding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallForwarding{}, &FirewallForwardingList{})
}
