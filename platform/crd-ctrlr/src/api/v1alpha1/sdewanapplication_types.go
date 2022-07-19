// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SdewanApplicationSpec defines the desired state of SdewanApplication
type SdewanApplicationSpec struct {
	PodSelector  *metav1.LabelSelector `json:"podSelector,omitempty"`
	AppNamespace string                `json:"appNamespace,omitempty"`
	ServicePort  string                `json:"servicePort,omitempty"`
	CNFPort      string                `json:"cnfPort,omitempty"`
}

type ApplicationInfo struct {
	IpList string
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SdewanApplication is the Schema for the sdewanapplications API
type SdewanApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    SdewanApplicationSpec `json:"spec,omitempty"`
	Status  SdewanStatus          `json:"status,omitempty"`
	AppInfo ApplicationInfo       `json:"-"`
}

// +kubebuilder:object:root=true

// SdewanApplicationList contains a list of SdewanApplication
type SdewanApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SdewanApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SdewanApplication{}, &SdewanApplicationList{})
}
