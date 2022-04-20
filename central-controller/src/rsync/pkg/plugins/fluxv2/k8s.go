// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	gitsupport "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/gitsupport"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

// Connection is for a cluster
type Fluxv2Provider struct {
	gitProvider gitsupport.GitProvider
}

func NewFluxv2Provider(cid, app, cluster, level, namespace string) (*Fluxv2Provider, error) {

	c, err := utils.GetGitOpsConfig(cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	if c.Props.GitOpsType != "fluxcd" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Props.GitOpsType)
	}

	gitProvider, err := gitsupport.NewGitProvider(cid, app, cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	p := Fluxv2Provider{
		gitProvider: *gitProvider,
	}
	return &p, nil
}

func (p *Fluxv2Provider) CleanClientProvider() error {
	return nil
}
