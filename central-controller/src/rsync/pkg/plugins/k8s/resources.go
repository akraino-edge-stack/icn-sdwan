// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8s

import (
	"context"
	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
)

// Creates a new resource if the not already existing
func (p *K8sProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	if err := p.client.Create(content); err != nil {
		if apierrors.IsAlreadyExists(err) {
			log.Warn("Resource is already present, Skipping", log.Fields{"error": err, "resource": name})
			return nil, nil
		} else {
			log.Error("Failed to create res", log.Fields{"error": err, "resource": name})
		}
		return nil, err
	}
	return nil, nil
}

// Apply resource to the cluster
func (p *K8sProvider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	if err := p.client.Apply(content); err != nil {
		log.Error("Failed to apply res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	acUtils, err := utils.NewAppContextReference(p.cid)
	if err != nil {
		return nil, nil
	}
	// Currently only subresource supported is approval
	subres, _, err := acUtils.GetSubResApprove(name, p.app, p.cluster)
	if err == nil {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return nil, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		log.Info("Approval Subresource::", log.Fields{"cluster": p.cluster, "resource": result[0], "approval": string(subres)})
		err = p.client.Approve(result[0], subres)
		return nil, err
	}

	return nil, nil
}

// Delete resource from the cluster
func (p *K8sProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {
	if err := p.client.Delete(content); err != nil {
		log.Error("Failed to delete res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	return nil, nil
}

// Get resource from the cluster
func (p *K8sProvider) Get(name string, gvkRes []byte) ([]byte, error) {
	b, err := p.client.Get(gvkRes, p.namespace)
	if err != nil {
		log.Error("Failed to get res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	return b, nil
}

// Commit resources to the cluster
// Not required in K8s case
func (p *K8sProvider) Commit(ctx context.Context, ref interface{}) error {
	return nil
}

// IsReachable cluster reachablity test
func (p *K8sProvider) IsReachable() error {
	return p.client.IsReachable()
}

func (m *K8sProvider) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
