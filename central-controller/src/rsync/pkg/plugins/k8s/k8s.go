// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8s

import (
	//"fmt"
	"io/ioutil"
	"os"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	kubeclient "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

// Connection is for a cluster
type K8sProvider struct {
	cid       string
	cluster   string
	app       string
	namespace string
	level     string
	fileName  string
	client    *kubeclient.Client
}

func NewK8sProvider(cid, app, cluster, level, namespace string) (*K8sProvider, error) {
	p := K8sProvider{
		cid:       cid,
		app:       app,
		cluster:   cluster,
		level:     level,
		namespace: namespace,
	}
	// Get file from DB
	dec, err := utils.GetKubeConfig(cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	//
	f, err := ioutil.TempFile("/tmp", "rsync-config-"+cluster+"-")
	if err != nil {
		log.Error("Unable to create temp file in tmp directory", log.Fields{"err": err})
		return nil, err
	}
	fileName := f.Name()
	log.Info("Temp file for Kubeconfig", log.Fields{"fileName": fileName})
	_, err = f.Write(dec)
	if err != nil {
		log.Error("Unable to write tmp directory", log.Fields{"err": err, "filename": fileName})
		return nil, err
	}

	client := kubeclient.New("", fileName, namespace)
	if client == nil {
		return nil, pkgerrors.New("failed to connect with the cluster")
	}
	p.fileName = fileName
	p.client = client
	return &p, nil
}

// If file exists delete it
func (p *K8sProvider) CleanClientProvider() error {
	if _, err := os.Stat(p.fileName); err == nil {
		os.Remove(p.fileName)
	}
	return nil

}
