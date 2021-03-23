// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"

	pkgerrors "github.com/pkg/errors"
)

type clientDbInfo struct {
	storeName    string // name of the mongodb collection to use for client documents
	tagNamespace string // attribute key name for the namespace section of a CloudConfig
	tagConfig    string // attribute key name for the kubeconfig section of a CloudConfig
}

// ClusterKey is the key structure that is used in the database
type CloudConfigKey struct {
	Provider  string `json:"provider"`
	Cluster   string `json:"cluster"`
	Level     string `json:"level"`
	Namespace string `json:"namespace"`
}

// CloudConfig contains the parameters that specify access to a cloud at any level
type CloudConfig struct {
	Provider  string `json:"provider"`
	Cluster   string `json:"cluster"`
	Level     string `json:"level"`
	Namespace string `json:"namespace"`
	Config    string `json:"config"`
}

type KubeConfig struct {
	Config string `json:"config"`
}

// CloudConfigManager is an interface that exposes the Cloud Config functionality
type CloudConfigManager interface {
	GetCloudConfig(provider string, cluster string, level string, namespace string) (CloudConfig, error)
	CreateCloudConfig(provider string, cluster string, level string, namespace string, config string) (CloudConfig, error)
	GetNamespace(provider string, cluster string) (CloudConfig, error)                   // level-0 only
	SetNamespace(provider string, cluster string, namespace string) (CloudConfig, error) // level-0 only
	DeleteCloudConfig(provider string, cluster string, level string, namespace string) error
}

// CloudConfigClient implements CloudConfigManager
// It will also be used to maintain some localized state
type CloudConfigClient struct {
	db clientDbInfo
}

// NewCloudConfigClient returns an instance of the CloudConfigClient
// which implements CloudConfigManager
func NewCloudConfigClient() *CloudConfigClient {
	return &CloudConfigClient{
		db: clientDbInfo{
			storeName:    "cloudconfig",
			tagNamespace: "namespace",
			tagConfig:    "config",
		},
	}
}

func unmarshal(values [][]byte) (KubeConfig, error) {
	if values != nil {
		kc := KubeConfig{}
		err := db.DBconn.Unmarshal(values[0], &kc)
		if err != nil {
			log.Error("Failed unmarshaling Value", log.Fields{})
			return KubeConfig{}, pkgerrors.Wrap(err, "Failed unmarshaling Value")
		}
		return kc, nil
	}
	log.Info("DB values is nil", log.Fields{})
	return KubeConfig{}, pkgerrors.New("DB values is nil")
}

// GetCloudConfig allows to get an existing cloud config entry
func (c *CloudConfigClient) GetCloudConfig(provider string, cluster string, level string, namespace string) (CloudConfig, error) {

	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagConfig)
	if err != nil {
		log.Error("Finding CloudConfig failed", log.Fields{})
		return CloudConfig{}, pkgerrors.Wrap(err, "Finding CloudConfig failed")
	}

	kc, err := unmarshal(values)
	if err != nil {
		return CloudConfig{}, err
	}

	cc := CloudConfig{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		Config:    kc.Config,
	}

	return cc, nil
}

// CreateCloudConfig allows to create a new cloud config entry to hold a kubeconfig for access
func (c *CloudConfigClient) CreateCloudConfig(provider string, cluster string, level string, namespace string, config string) (CloudConfig, error) {

	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	kc := KubeConfig{
		Config: config,
	}

	// check if it already exists
	_, err := c.GetCloudConfig(provider, cluster, level, namespace)
	if err == nil {
		log.Error("CloudConfig already exists", log.Fields{})
		return CloudConfig{}, pkgerrors.New("CloudConfig already exists")
	}

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagConfig, kc)
	if err != nil {
		log.Error("Failure inserting CloudConfig", log.Fields{})
		return CloudConfig{}, pkgerrors.Wrap(err, "Failure inserting CloudConfig")
	}

	cc := CloudConfig{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		Config:    config,
	}

	return cc, nil
}

// SetNamespace is only for L0 cloud configs and allows to set/reset current namespace name
func (c *CloudConfigClient) SetNamespace(provider string, cluster string, namespace string) error {
	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     "0", // always going to be level 0 if we're (un)setting a namespace name
		Namespace: "",  // we don't care about the current name for now
	}

	// check if CloudConfig exists and also get the current namespace name
	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagNamespace)
	if err != nil {
		log.Error("Could not fetch the CloudConfig so not updating", log.Fields{})
		return pkgerrors.Wrap(err, "Could not fetch the CloudConfig so not updating")
	}

	newkey := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     "0",               // always going to be level 0 if we're (un)setting a namespace name
		Namespace: string(values[0]), // use current namespace name as the final key to update the namespace
	}
	err = db.DBconn.Insert(c.db.storeName, newkey, nil, c.db.tagNamespace, namespace)
	if err != nil {
		log.Error("Could not update the namespace of the CloudConfig", log.Fields{})
		return pkgerrors.Wrap(err, "Could not update the namespace of the CloudConfig")
	}

	return nil
}

// GetNamespace is only for L0 cloud configs and allows fetching the current namespace name
func (c *CloudConfigClient) GetNamespace(provider string, cluster string) (string, error) {
	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     "0", // always going to be level 0 if getting a namespace name without ambiguity
		Namespace: "",  // we don't care about the current name
	}

	// check if CloudConfig exists and also get the current namespace name
	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagNamespace)
	if err != nil {
		log.Error("Could not fetch the CloudConfig so can't return namespace", log.Fields{})
		return "", pkgerrors.Wrap(err, "Could not fetch the CloudConfig so can't return namespace")
	}
	if len(values) > 1 {
		log.Error("Multiple CloudConfigs were returned, which was unexpected", log.Fields{})
		return "", pkgerrors.New("Multiple CloudConfigs were returned, which was unexpected")
	}
	if len(values) == 0 {
		log.Error("No CloudConfig was returned", log.Fields{})
		return "", pkgerrors.New("No CloudConfig was returned") // error message is evaluated at calling function, don't change
	}

	return string(values[0]), nil
}

// DeleteCloudConfig deletes a cloud config entry
func (c *CloudConfigClient) DeleteCloudConfig(provider string, cluster string, level string, namespace string) error {
	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	// check if it doesn't exist
	_, err := c.GetCloudConfig(provider, cluster, level, namespace)
	if err != nil {
		log.Error("Could not fetch the CloudConfig so not deleting", log.Fields{})
		return pkgerrors.New("Could not fetch the CloudConfig so not deleting")
	}

	err = db.DBconn.Remove(c.db.storeName, key)
	if err != nil {
		log.Error("Could not delete the CloudConfig", log.Fields{})
		return pkgerrors.Wrap(err, "Could not delete the CloudConfig")
	}

	return nil
}
