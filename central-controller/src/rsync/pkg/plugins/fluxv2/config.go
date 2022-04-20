// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package fluxv2

import (
	"context"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomize "github.com/fluxcd/kustomize-controller/api/v1beta2"
	fluxsc "github.com/fluxcd/source-controller/api/v1beta1"
	yaml "github.com/ghodss/yaml"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) ApplyConfig(ctx context.Context, config interface{}) error {

	// Create Source CR and Kcustomize CR
	gr := fluxsc.GitRepository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "source.toolkit.fluxcd.io/v1beta1",
			Kind:       "GitRepository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.gitProvider.Cid,
			Namespace: p.gitProvider.Namespace,
		},
		Spec: fluxsc.GitRepositorySpec{
			URL:       p.gitProvider.Url,
			Interval:  metav1.Duration{Duration: time.Second * 30},
			Reference: &fluxsc.GitRepositoryRef{Branch: p.gitProvider.Branch},
		},
	}
	x, err := yaml.Marshal(&gr)
	if err != nil {
		log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "gr": gr})
		return err
	}
	path := "clusters/" + p.gitProvider.Cluster + "/" + gr.Name + ".yaml"
	// Add to the commit
	gp := emcogit.Add(path, string(x), []gitprovider.CommitFile{}, p.gitProvider.GitType)

	kc := kustomize.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kustomize.toolkit.fluxcd.io/v1beta2",
			Kind:       "Kustomization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kcust" + p.gitProvider.Cid,
			Namespace: p.gitProvider.Namespace,
		},
		Spec: kustomize.KustomizationSpec{
			Interval: metav1.Duration{Duration: time.Second * 300},
			Path:     "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid,
			Prune:    true,
			SourceRef: kustomize.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: gr.Name,
			},
			TargetNamespace: p.gitProvider.Namespace,
		},
	}
	y, err := yaml.Marshal(&kc)
	if err != nil {
		log.Error("ApplyConfig:: Marshal err", log.Fields{"err": err, "kc": kc})
		return err
	}
	path = "clusters/" + p.gitProvider.Cluster + "/" + kc.Name + ".yaml"
	gp = emcogit.Add(path, string(y), gp, p.gitProvider.GitType)
	// Commit
	err = emcogit.CommitFiles(ctx, p.gitProvider.Client, p.gitProvider.UserName, p.gitProvider.RepoName, p.gitProvider.Branch, "Commit for "+p.gitProvider.GetPath("context"), gp, p.gitProvider.GitType)
	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
	}
	return err
}

// Delete GitRepository and Kustomization CR's for Flux
func (p *Fluxv2Provider) DeleteConfig(ctx context.Context, config interface{}) error {
	path := "clusters/" + p.gitProvider.Cluster + "/" + p.gitProvider.Cid + ".yaml"
	gp := emcogit.Delete(path, []gitprovider.CommitFile{}, p.gitProvider.GitType)
	path = "clusters/" + p.gitProvider.Cluster + "/" + "kcust" + p.gitProvider.Cid + ".yaml"
	gp = emcogit.Delete(path, gp, p.gitProvider.GitType)
	err := emcogit.CommitFiles(ctx, p.gitProvider.Client, p.gitProvider.UserName, p.gitProvider.RepoName, p.gitProvider.Branch, "Commit for "+p.gitProvider.GetPath("context"), gp, p.gitProvider.GitType)
	if err != nil {
		log.Error("ApplyConfig:: Commit files err", log.Fields{"err": err, "gp": gp})
	}
	return err
}
