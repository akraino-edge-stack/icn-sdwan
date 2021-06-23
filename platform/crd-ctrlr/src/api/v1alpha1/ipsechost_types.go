// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Connection struct {
	Name           string   `json:"name"`
	ConnectionType string   `json:"conn_type"`
	Mode           string   `json:"mode"`
	LocalSourceIp  string   `json:"local_sourceip,omitempty"`
	LocalUpDown    string   `json:"local_updown,omitempty"`
	LocalFirewall  string   `json:"local_firewall,omitempty"`
	RemoteSubnet   string   `json:"remote_subnet,omitempty"`
	RemoteSourceIp string   `json:"remote_sourceip,omitempty"`
	RemoteUpDown   string   `json:"remote_updown,omitempty"`
	RemoteFirewall string   `json:"remote_firewall,omitempty"`
	CryptoProposal []string `json:"crypto_proposal,omitempty"`
	Mark           string   `json:"mark,omitempty"`
	IfId           string   `json:"if_id,omitempty"`
}

type IpsecHostSpec struct {
	Name                 string       `json:"name,omitempty"`
	Type                 string       `json:"type,omitempty"`
	Remote               string       `json:"remote"`
	AuthenticationMethod string       `json:"authentication_method"`
	CryptoProposal       []string     `json:"crypto_proposal"`
	LocalIdentifier      string       `json:"local_identifier,omitempty"`
	RemoteIdentifier     string       `json:"remote_identifier,omitempty"`
	ForceCryptoProposal  string       `json:"force_crypto_proposal,omitempty"`
	PresharedKey         string       `json:"pre_shared_key,omitempty"`
	LocalPublicCert      string       `json:"local_public_cert,omitempty"`
	LocalPrivateCert     string       `json:"local_private_cert,omitempty"`
	SharedCA             string       `json:"shared_ca,omitempty"`
	Connections          []Connection `json:"connections"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IpsecHost is the Schema for the ipsechosts API
type IpsecHost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpsecHostSpec `json:"spec,omitempty"`
	Status SdewanStatus  `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IpsecHostList contains a list of IpsecHost
type IpsecHostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpsecHost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpsecHost{}, &IpsecHostList{})
}
