// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
)

// Creates a new resource if the not already existing
func (p *AzureArcProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *AzureArcProvider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Apply(name, ref, content)
	return res, err
}

// Delete resource from the cluster
func (p *AzureArcProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Delete(name, ref, content)
	return res, err
}

// Get resource from the cluster
func (p *AzureArcProvider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *AzureArcProvider) Commit(ctx context.Context, ref interface{}) error {

	err := p.gitProvider.Commit(ctx, ref)
	return err
}

// IsReachable cluster reachablity test
func (p *AzureArcProvider) IsReachable() error {
	return nil
}

func (p *AzureArcProvider) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
