// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"

)

// StartClusterWatcher watches for CR changes in git location
func (p *Fluxv2Provider) StartClusterWatcher() error {
	p.gitProvider.StartClusterWatcher()
	return nil
}

// ApplyStatusCR applies status CR
func (p *Fluxv2Provider) ApplyStatusCR(name string, content []byte) error {

	// Add namespace to the status resource, needed by
	// Flux
	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(content), unstruct)
	if err != nil {
		return err
	}
	// Set Namespace
	unstruct.SetNamespace(p.gitProvider.Namespace)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return err
	}
	ref, err := p.gitProvider.Apply(name, nil, b)
	if err != nil {
		return err
	}
	p.gitProvider.Commit(context.Background(), ref)
	return err

}

// DeleteStatusCR deletes status CR
func (p *Fluxv2Provider) DeleteStatusCR(name string, content []byte) error {
	ref, err := p.gitProvider.Delete(name, nil, content)
	if err != nil {
		return err
	}
	p.gitProvider.Commit(context.Background(), ref)
	return err
}
