package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SdewanSpec defines the desired state of Sdewan
type Network struct {
	Name string  `json:"name"`

	// +optional
	IsProvider bool  `json:"isProvider"`

	// +optional
	Interface string  `json:"interface"`

	// +optional
	DefaultGateway bool  `json:"defaultGateway"`
}

// SdewanSpec defines the desired state of Sdewan
type SdewanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Node string  `json:"node"`
	Networks []Network  `json:"networks"`

	// +optional
	Mwan3Rule string  `json:"mwan3Rule"`
}

// SdewanStatus defines the observed state of Sdewan
type SdewanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sdewan is the Schema for the sdewans API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sdewans,scope=Namespaced
type Sdewan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SdewanSpec   `json:"spec,omitempty"`
	Status SdewanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SdewanList contains a list of Sdewan
type SdewanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sdewan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sdewan{}, &SdewanList{})
}
