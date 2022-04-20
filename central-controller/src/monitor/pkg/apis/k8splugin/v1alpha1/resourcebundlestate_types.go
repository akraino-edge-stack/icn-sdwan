/*
Copyright 2022.

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
// +k8s:deepcopy-gen=package,register
// +groupName=k8splugin.io
package v1alpha1

// +kubebuilder:validation:Optional
import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceBundleStateSpec defines the desired state of ResourceBundleState
type ResourceBundleStateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`
}

// ResourceBundleStateStatus defines the observed state of ResourceBundleState
type ResourceBundleStateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready         bool  `json:"ready" protobuf:"varint,1,opt,name=ready"`
	ResourceCount int32 `json:"resourceCount" protobuf:"varint,2,opt,name=resourceCount"`

	// +kubebuilder:validation:Optional
	// +optional
	PodStatuses []corev1.Pod `json:"podStatuses,omitempty" protobuf:"varint,3,opt,name=podStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	ServiceStatuses []corev1.Service `json:"serviceStatuses,omitempty" protobuf:"varint,4,opt,name=serviceStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	ConfigMapStatuses []corev1.ConfigMap `json:"configMapStatuses,omitempty" protobuf:"varint,5,opt,name=configMapStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	DeploymentStatuses []appsv1.Deployment `json:"deploymentStatuses,omitempty" protobuf:"varint,6,opt,name=deploymentStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	DaemonSetStatuses []appsv1.DaemonSet `json:"daemonSetStatuses,omitempty" protobuf:"varint,8,opt,name=daemonSetStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	CsrStatuses []certsapi.CertificateSigningRequest `json:"csrStatuses,omitempty" protobuf:"varint,9,opt,name=csrStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	JobStatuses []v1.Job `json:"jobStatuses,omitempty" protobuf:"varint,12,opt,name=jobStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	StatefulSetStatuses []appsv1.StatefulSet `json:"statefulSetStatuses,omitempty" protobuf:"varint,13,opt,name=statefulSetStatuses"`

	// +kubebuilder:validation:Optional
	// +optional
	ResourceStatuses []ResourceStatus `json:"resourceStatuses,omitempty" protobuf:"varint,14,opt,name=resourceStatuses"`
}

// Resoureces
type ResourceStatus struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Res       []byte `json:"res"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// ResourceBundleState is the Schema for the resourcebundlestates API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

type ResourceBundleState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceBundleStateSpec   `json:"spec,omitempty"`
	Status ResourceBundleStateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceBundleStateList contains a list of ResourceBundleState
type ResourceBundleStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceBundleState `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceBundleState{}, &ResourceBundleStateList{})
}
