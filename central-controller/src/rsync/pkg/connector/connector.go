// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package connector

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	kubeclient "github.com/open-ness/EMCO/src/rsync/pkg/client"
	"github.com/open-ness/EMCO/src/rsync/pkg/db"
	pkgerrors "github.com/pkg/errors"
)

type Connector struct {
	cid     string
	Clients map[string]*kubeclient.Client
	sync.Mutex
}

const basePath string = "/tmp/rsync/"

// Init connector for an app context
func Init(id interface{}) *Connector {
	c := make(map[string]*kubeclient.Client)
	str := fmt.Sprintf("%v", id)
	return &Connector{
		Clients: c,
		cid:     str,
	}
}

// GetKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
func GetKubeConfig(clustername string, level string, namespace string) ([]byte, error) {
	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}

	ccc := db.NewCloudConfigClient()

	cconfig, err := ccc.GetCloudConfig(strs[0], strs[1], level, namespace)
	if err != nil {
		return nil, pkgerrors.New("Get kubeconfig failed")
	}
	log.Info("Successfully looked up CloudConfig", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

	dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// GetClient returns client for the cluster
func (c *Connector) GetClient(cluster string, level string, namespace string) (*kubeclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	client, ok := c.Clients[cluster]
	if !ok {
		// Get file from DB
		dec, err := GetKubeConfig(cluster, level, namespace)
		if err != nil {
			return nil, err
		}
		var kubeConfigPath string = basePath + c.cid + "/" + cluster + "/"
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
		}
	}
	return client, nil
}

func (c *Connector) GetClientWithRetry(cluster string, level string, namespace string) (*kubeclient.Client, error) {
	client, err := c.GetClient(cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	if err = client.IsReachable(); err != nil {
		return nil, err // TODO: Add retry
	}
	return client, nil
}

func (c *Connector) RemoveClient() {
	c.Lock()
	defer c.Unlock()
	err := os.RemoveAll(basePath + "/" + c.cid)
	if err != nil {
		log.Error("Warning: Deleting kubepath", log.Fields{"err": err})
	}
}
