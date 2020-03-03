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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type FirewallZone struct {
	Name string `json:"name"`
	// +optional
	Network []string `json:"network"`
	// +optional
	Masq string `json:"masq"`
	// +optional
	MasqSrc []string `json:"masq_src"`
	// +optional
	MasqDest []string `json:"masq_dest"`
	// +optional
	MasqAllowInvalid string `json:"masq_allow_invalid"`
	// +optional
	MtuFix string `json:"mtu_fix"`
	// +optional
	Input string `json:"input"`
	// +optional
	Forward string `json:"forward"`
	// +optional
	Output string `json:"output"`
	// +optional
	Family string `json:"family"`
	// +optional
	Subnet []string `json:"subnet"`
	// +optional
	ExtraSrc string `json:"extra_src"`
	// +optional
	ExtraDest string `json:"etra_dest"`
}

type FirewallRedirect struct {
	Name string `json:"name"`
	// +optional
	Src string `json:"src"`
	// +optional
	SrcIp string `json:"src_ip"`
	// +optional
	SrcDIp string `json:"src_dip"`
	// +optional
	SrcMac string `json:"src_mac"`
	// +optional
	SrcPort string `json:"src_port"`
	// +optional
	SrcDPort string `json:"src_dport"`
	// +optional
	Proto string `json:"proto"`
	// +optional
	Dest string `json:"dest"`
	// +optional
	DestIp string `json:"dest_ip"`
	// +optional
	DestPort string `json:"dest_port"`
	// +optional
	Mark string `json:"mark"`
	// +optional
	Target string `json:"target"`
	// +optional
	Family string `json:"family"`
}

type FirewallRule struct {
	Name string `json:"name"`
	// +optional
	Src string `json:"src"`
	// +optional
	SrcIp string `json:"src_ip"`
	// +optional
	SrcMac string `json:"src_mac"`
	// +optional
	SrcPort string `json:"src_port"`
	// +optional
	Proto string `json:"proto"`
	// +optional
	IcmpType []string `json:"icmp_type"`
	// +optional
	Dest string `json:"dest"`
	// +optional
	DestIp string `json:"dest_ip"`
	// +optional
	DestPort string `json:"dest_port"`
	// +optional
	Mark string `json:"mark"`
	// +optional
	Target string `json:"target"`
	// +optional
	SetMark string `json:"set_mark"`
	// +optional
	SetXmark string `json:"set_xmark"`
	// +optional
	Family string `json:"family"`
	// +optional
	Extra string `json:"extra"`
}

type FirewallForwarding struct {
	Name string `json:"name"`
	// +optional
	Src string `json:"src"`
	// +optional
	Dest string `json:"dest"`
	// +optional
	Family string `json:"family"`
}

// FirewallConfSpec defines the desired state of FirewallConf
type FirewallConfSpec struct {
	// Foo is an example field of FirewallConf. Edit FirewallConf_types.go to remove/update
	Zones       []FirewallZone       `json:"zones"`
	Redirects   []FirewallRedirect   `json:"redirects"`
	Rules       []FirewallRule       `json:"rules"`
	Forwardings []FirewallForwarding `json:"forwardings"`
}

// FirewallConfStatus defines the observed state of FirewallConf
type FirewallConfStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// FirewallConf is the Schema for the firewallconfs API
type FirewallConf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallConfSpec   `json:"spec,omitempty"`
	Status FirewallConfStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FirewallConfList contains a list of FirewallConf
type FirewallConfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallConf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallConf{}, &FirewallConfList{})
}
