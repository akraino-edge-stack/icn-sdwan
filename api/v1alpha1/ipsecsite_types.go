/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SiteConnection struct {
	Name string `json:"name"`

	ConnectionType string `json:"type"`

	Mode string `json:"mode"`

	//+optional
	LocalSubnet string `json:"local_subnet"`

	//+optional
	LocalNAT string `json:"local_NAT"`

	//+optional
	LocalSourceIp string `json:"local_sourceip"`

	//+optional
	LocalUpDown string `json:"local_updown"`

	//+optional
	LocalFirewall string `json:"local_firewall"`

	//+optional
	RemoteSubnet string `json:"remote_subnet"`

	//+optional
	RemoteSourceIp string `json:"remote_sourceip"`

	//+optional
	RemoteUpDown string `json:"remote_updown"`

	//+optional
	RemoteFirewall string `json:"remote_firewall"`

	CryptoProposal []string `json:"crypto_proposal"`
}

type SiteConnections struct {
	SiteName string `json:"name"`

	Gateway string `json:"gateway"`

	AuthenticationMethod string `json:"authentication_method"`

	CryptoProposal []string `json:"crypto_proposal"`

	//+optional
	LocalIdentifier string `json:"local_identifier"`

	//+optional
	RemoteIdentifier string `json:"remote_identifier"`

	//+optional
	ForceCryptoProposal bool `json:"force_crypto_proposal"`

	//+optional
	PresharedKey string `json:"pre_shared_key"`

	//+optional
	LocalPublicCert string `json:"local_public_cert"`

	//+optional
	LocalPrivateCert string `json:"local_private_cert"`

	//+optional
	SharedCA string `json:"shared_ca"`

	Connections []SiteConnection `json:"connections"`
}

type CryptoProposal struct {
	Name                string `json:"name"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	HashAlgorithm       string `json:"hash_algorithm"`
	DhGroup             string `json:"dh_group"`
}

// IpsecSiteSpec defines the desired state of IpsecSite
type IpsecSiteSpec struct {
	Site     SiteConnections  `json:"site"`
	Proposal []CryptoProposal `json:"proposal"`
}

// IpsecSiteStatus defines the observed state of IpsecSite
type IpsecSiteStatus struct {

	//+optional
	Active []corev1.ObjectReference `json:"active,omitempty"`
}

// +kubebuilder:object:root=true

// IpsecSite is the Schema for the ipsecsites API
type IpsecSite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IpsecSiteSpec   `json:"spec,omitempty"`
	Status IpsecSiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IpsecSiteList contains a list of IpsecSite
type IpsecSiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IpsecSite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IpsecSite{}, &IpsecSiteList{})
}
