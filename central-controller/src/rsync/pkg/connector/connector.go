// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package connector

import (
	"errors"
	"fmt"
	"os"
	"sync"
	//types "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	kubeclient "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/client"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

// IsTestKubeClient .. global variable used during unit-tests to check whether a fake kube client object has to be instantiated
var IsTestKubeClient bool = false

// Connection is for a cluster
type Connection struct {
	Cid     string
	Clients map[string]*kubeclient.Client
	sync.Mutex
}

const basePath string = "/tmp/rsync/"

// Init Connection for an app context
func (c *Connection) Init(id interface{}) error {
	log.Info("Init with interface", log.Fields{})
	c.Clients = make(map[string]*kubeclient.Client)
	c.Cid = fmt.Sprintf("%v", id)
	return nil
}

// GetClient returns client for the cluster
func (c *Connection) GetClient(cluster string, level string, namespace string) (*kubeclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	// Check if Fake kube client is required(it's true for unit tests)
	log.Info("GetClient .. start", log.Fields{"IsTestKubeClient": fmt.Sprintf("%v", IsTestKubeClient)})
	if IsTestKubeClient {
		return kubeclient.NewKubeFakeClient()
	}

	client, ok := c.Clients[cluster]
	if !ok {
		// Get file from DB
		dec, err := utils.GetKubeConfig(cluster, level, namespace)
		if err != nil {
			return nil, err
		}
		var kubeConfigPath string = basePath + c.Cid + "/" + cluster + "/"
		if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
			err = os.MkdirAll(kubeConfigPath, 0700)
			if err != nil {
				return nil, err
			}
		}
		kubeConfig := kubeConfigPath + "config"
		f, err := os.Create(kubeConfig)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(dec)
		if err != nil {
			return nil, err
		}
		client = kubeclient.New("", kubeConfig, namespace)
		if client != nil {
			c.Clients[cluster] = client
		} else {
			return nil, errors.New("failed to connect with the cluster")
		}
	}
	return client, nil
}

func (c *Connection) RemoveClient() {
	c.Lock()
	defer c.Unlock()
	err := os.RemoveAll(basePath + "/" + c.Cid)
	if err != nil {
		log.Error("Warning: Deleting kubepath", log.Fields{"err": err})
	}
}
