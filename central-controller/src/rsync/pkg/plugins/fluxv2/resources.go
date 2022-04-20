// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"

)

// Creates a new resource if the not already existing
func (p *Fluxv2Provider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Create(name, ref, content)
	return res, err
}

// Apply resource to the cluster
func (p *Fluxv2Provider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return nil, err
	}
	// Set Namespace
	unstruct.SetNamespace(p.gitProvider.Namespace)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}
	res, err := p.gitProvider.Apply(name, ref, b)
	return res, err

}

// Delete resource from the cluster
func (p *Fluxv2Provider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	res, err := p.gitProvider.Delete(name, ref, content)
	return res, err

}

// Get resource from the cluster
func (p *Fluxv2Provider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

// Commit resources to the cluster
func (p *Fluxv2Provider) Commit(ctx context.Context, ref interface{}) error {

	err := p.gitProvider.Commit(ctx, ref)
	return err
}

// IsReachable cluster reachablity test
func (p *Fluxv2Provider) IsReachable() error {
	return nil
}

func (m *Fluxv2Provider) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}

