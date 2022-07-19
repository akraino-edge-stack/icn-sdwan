// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CNFHubSiteStatus defines the observed state of CNFHubSiteStatus
type CNFHubSiteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Type string `json:"type,omitempty"`
	// +optional
	SiteIPs []string `json:"remoteips,omitempty"`
	// +optional
	Subnet string `json:"subnet,omitempty"`
	// +optional
	HubIP string `json:"hubip,omitempty"`
	// +optional
	DevicePIP string `json:"devicepip,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

func (c *CNFHubSiteStatus) IsEqual(s *CNFHubSiteStatus) bool {
	if c.Type != s.Type ||
		c.Subnet != s.Subnet ||
		c.HubIP != s.HubIP ||
		c.DevicePIP != s.DevicePIP {
		return false
	}

	if len(c.SiteIPs) != len(s.SiteIPs) {
		return false
	}

	for i := 0; i < len(c.SiteIPs); i++ {
		if c.SiteIPs[i] != s.SiteIPs[i] {
			return false
		}
	}

	return true
}

// CNFHubSiteSpec defines the desired state of CNFHubSite
type CNFHubSiteSpec struct {
	Type      string `json:"type,omitempty"`
	Site      string `json:"site,omitempty"`
	Subnet    string `json:"subnet,omitempty"`
	HubIP     string `json:"hubip,omitempty"`
	DevicePIP string `json:"devicepip,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CNFHubSite is the Schema for the cnfhubsites API
type CNFHubSite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CNFHubSiteSpec   `json:"spec,omitempty"`
	Status CNFHubSiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CNFHubSiteList contains a list of CNFHubSite
type CNFHubSiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CNFHubSite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CNFHubSite{}, &CNFHubSiteList{})
}
