// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FirewallDNATSpec defines the desired state of FirewallDNAT
type FirewallDNATSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name     string `json:"name,omitempty"`
	Src      string `json:"src,omitempty"`
	SrcIp    string `json:"src_ip,omitempty"`
	SrcDIp   string `json:"src_dip,omitempty"`
	SrcMac   string `json:"src_mac,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	SrcDPort string `json:"src_dport,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Dest     string `json:"dest,omitempty"`
	DestIp   string `json:"dest_ip,omitempty"`
	DestPort string `json:"dest_port,omitempty"`
	Mark     string `json:"mark,omitempty"`
	Target   string `json:"target,omitempty"`
	Family   string `json:"family,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FirewallDNAT is the Schema for the firewalldnats API
type FirewallDNAT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallDNATSpec `json:"spec,omitempty"`
	Status SdewanStatus     `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallDNATList contains a list of FirewallDNAT
type FirewallDNATList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallDNAT `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallDNAT{}, &FirewallDNATList{})
}
