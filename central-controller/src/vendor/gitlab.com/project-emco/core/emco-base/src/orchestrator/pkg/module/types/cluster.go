// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package types

// GitOps shared objects between controllers - clm and rsync

// GitOps Spec
type GitOpsSpec struct {
	Props GitOpsProps `json:"gitOps"`
}
// GitOps Properties for Reference and Resource Objects
type GitOpsProps struct {
	// GitOps type - example Fluxv2, AzureArc, GoogleAnthos
	GitOpsType            string `json:"gitOpsType"`
	// Refrence Sync object for the cloud configuration
	GitOpsReferenceObject string `json:"gitOpsReferenceObject"`
	// Resource Sync Object for resurces
	GitOpsResourceObject  string `json:"gitOpsResourceObject"`
}
// Sync Objects
type ClusterSyncObjects struct {
	Metadata Metadata       `json:"metadata"`
	Spec     ClusterSyncObjectSpec `json:"spec"`
}
// Key value pairs for Sync Objects
type ClusterSyncObjectSpec struct {
	Kv []map[string]interface{} `json:"kv" encrypted:""`
}