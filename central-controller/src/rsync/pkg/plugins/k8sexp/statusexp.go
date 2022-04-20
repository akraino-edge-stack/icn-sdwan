// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8sexp

import (
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/plugins/k8s"
)

// StartClusterWatcher watches for CR
// Same as K8s
func (c *K8sProviderExp) StartClusterWatcher() error {
	cl, err := k8s.NewK8sProvider(c.cid, c.app, c.cluster, c.level, c.namespace)
	defer cl.CleanClientProvider()
	if err != nil {
		return err
	}
	return cl.StartClusterWatcher()
}

// ApplyStatusCR applies status CR
func (p *K8sProviderExp) ApplyStatusCR(name string, content []byte) error {
	if err := p.client.Apply(content); err != nil {
		log.Error("Failed to apply Status CR", log.Fields{
			"error": err,
		})
		return err
	}
	return nil

}

// DeleteStatusCR deletes status CR
func (p *K8sProviderExp) DeleteStatusCR(name string, content []byte) error {
	if err := p.client.Delete(content); err != nil {
		log.Error("Failed to delete Status CR", log.Fields{
			"error": err,
		})
		return err
	}
	return nil
}
