// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resourcestatus

// ResourceStatus struct is used to maintain the rsync status for resources in the appcontext
// that rsync is synchronizing to clusters
type ResourceStatus struct {
	Status RsyncStatus
}

type RsyncStatus = string

type statusValues struct {
	Pending  RsyncStatus
	Applied  RsyncStatus
	Failed   RsyncStatus
	Retrying RsyncStatus
	Deleted  RsyncStatus
}

var RsyncStatusEnum = &statusValues{
	Pending:  "Pending",
	Applied:  "Applied",
	Failed:   "Failed",
	Retrying: "Retrying",
	Deleted:  "Deleted",
}
