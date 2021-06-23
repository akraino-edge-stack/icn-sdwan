// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SdewanState string

const (
	InSync   SdewanState = "In Sync"
	Idle     SdewanState = "Idle"
	Applying SdewanState = "Trying to apply"
	Deleting SdewanState = "Being delete"
	Unknown  SdewanState = "Unknown status"
)

// status subsource used for Sdewan rule CRDs
type SdewanStatus struct {
	// +optional
	AppliedGeneration int64 `json:"appliedGeneration,omitempty"`
	// +optional
	AppliedTime *metav1.Time `json:"appliedTime,omitempty"`
	State       SdewanState  `json:"state"`
	// +optional
	Message string `json:"message,omitempty"`
}
