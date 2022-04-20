// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package client

import (
	"context"
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext/subresources"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	k8scertsv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) Approve(name string, sa []byte) error {

	var a subresources.ApprovalSubresource

	err := json.Unmarshal(sa, &a)
	if err != nil {
		return pkgerrors.Wrap(err, "An error occurred while parsing the approval Subresource.")
	}

	csr, err := c.Clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return pkgerrors.Wrap(err, "could not get the certificate")
	}
	var timePtr metav1.Time
	str := []string{a.LastUpdateTime}
	if err = metav1.Convert_Slice_string_To_v1_Time(&str, &timePtr, nil); err != nil {
		return pkgerrors.Wrap(err, "An error occurred while converting time from string.")
	}
	// Update CSR with Conditions
	csr.Status.Conditions = append(csr.Status.Conditions, k8scertsv1.CertificateSigningRequestCondition{
		Type:           k8scertsv1.RequestConditionType(a.Type),
		Reason:         a.Reason,
		Message:        a.Message,
		LastUpdateTime: timePtr,
		Status:         a.Status,
	})
	// CSR Approval
	_, err = c.Clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), csr.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		logutils.Error("Failed to UpdateApproval", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	return nil
}
