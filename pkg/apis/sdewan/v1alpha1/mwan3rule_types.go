package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.


type PolicyMember struct {
        Network string `json:"network"`
	Metric int `json:"metric"`
	Weight int `json:"weight"`
}

type RuleItem struct {
        Name string `json:"name"`
	UsePolicy string `json:"use_policy"`
	DestIP string `json:"dest_ip"`
}
// Mwan3RuleSpec defines the desired state of Mwan3Rule
type Mwan3RuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Policy map[string][]PolicyMember  `json:"policy"`
	Rule []RuleItem `json:"rule"`
}

// Mwan3RuleStatus defines the observed state of Mwan3Rule
type Mwan3RuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mwan3Rule is the Schema for the mwan3rules API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=mwan3rules,scope=Namespaced
type Mwan3Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Mwan3RuleSpec   `json:"spec,omitempty"`
	Status Mwan3RuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Mwan3RuleList contains a list of Mwan3Rule
type Mwan3RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mwan3Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mwan3Rule{}, &Mwan3RuleList{})
}
