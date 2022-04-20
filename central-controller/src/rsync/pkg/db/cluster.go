// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"

	pkgerrors "github.com/pkg/errors"
)

type clientDbInfo struct {
	storeName    string // name of the mongodb collection to use for client documents
	tagNamespace string // attribute key name for the namespace section of a CloudConfig
	tagConfig    string // attribute key name for the kubeconfig section of a CloudConfig
	tagMeta      string // attribute key name for the GitOps related Objects
}

// ClusterKey is the key structure that is used in the database
type CloudConfigKey struct {
	Provider  string `json:"cloudConfigClusterProvider"`
	Cluster   string `json:"cloudConfigCluster"`
	Level     string `json:"level"`
	Namespace string `json:"namespace"`
}

// CloudConfig contains the parameters that specify access to a cloud at any level
type CloudConfig struct {
	Provider  string `json:"cloudConfigClusterProvider"`
	Cluster   string `json:"cloudConfigCluster"`
	Level     string `json:"level"`
	Namespace string `json:"namespace"`
	Config    string `json:"config"`
}

type KubeConfig struct {
	Config string `json:"config" encrypted:""`
}

// ClusterSyncObjectKey is the key structure that is used in the database
type ClusterSyncObjectsKey struct {
	ClusterProviderName    string `json:"clusterProvider"`
	ClusterSyncObjectsName string `json:"clusterSyncObject"`
}

// CloudConfig contains the parameters that specify access to a cloud at any level
type CloudGitOpsConfig struct {
	Provider  string            `json:"cloudConfigClusterProvider"`
	Cluster   string            `json:"cloudConfigCluster"`
	Level     string            `json:"level"`
	Namespace string            `json:"namespace"`
	Config    mtypes.GitOpsSpec `json:"gitOps"`
}

// CloudConfigManager is an interface that exposes the Cloud Config functionality
type CloudConfigManager interface {
	GetCloudConfig(provider string, cluster string, level string, namespace string) (CloudConfig, error)
	CreateCloudConfig(provider string, cluster string, level string, namespace string, config string) (CloudConfig, error)
	GetNamespace(provider string, cluster string) (string, error)         // level-0 only
	SetNamespace(provider string, cluster string, namespace string) error // level-0 only
	DeleteCloudConfig(provider string, cluster string, level string, namespace string) error
	CreateClusterSyncObjects(provider string, pr mtypes.ClusterSyncObjects, exists bool) (mtypes.ClusterSyncObjects, error)
	GetClusterSyncObjects(provider, syncobject string) (mtypes.ClusterSyncObjects, error)
	GetClusterSyncObjectsValue(provider, syncobject, syncobjectkey string) (interface{}, error)
	GetAllClusterSyncObjects(provider string) ([]mtypes.ClusterSyncObjects, error)
	DeleteClusterSyncObjects(provider, syncobject string) error
	CreateGitOpsConfig(provider string, cluster string, gs mtypes.GitOpsSpec, level string, namespace string) (CloudGitOpsConfig, error)
	GetGitOpsConfig(provider string, cluster string, level string, namespace string) (CloudGitOpsConfig, error)
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
			storeName:    "resources",
			tagNamespace: "namespace",
			tagConfig:    "config",
			tagMeta:      "meta",
		},
	}
}

func unmarshal(values [][]byte) (KubeConfig, error) {
	if values != nil {
		kc := KubeConfig{}
		err := db.DBconn.Unmarshal(values[0], &kc)
		if err != nil {
			log.Error("Failed unmarshalling Value", log.Fields{})
			return KubeConfig{}, err
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

func (c *CloudConfigClient) CreateClusterSyncObjects(provider string, p mtypes.ClusterSyncObjects, exists bool) (mtypes.ClusterSyncObjects, error) {
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: p.Metadata.Name,
	}

	//Check if this ClusterSyncObjects already exists
	_, err := c.GetClusterSyncObjects(provider, p.Metadata.Name)
	if err == nil && !exists {
		return mtypes.ClusterSyncObjects{}, pkgerrors.New("Cluster Sync Objects already exists")
	}

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, p)
	if err != nil {
		return mtypes.ClusterSyncObjects{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterSyncObjects returns the Cluster Sync objects for corresponding provider and sync object name
func (c *CloudConfigClient) GetClusterSyncObjects(provider, syncobject string) (mtypes.ClusterSyncObjects, error) {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return mtypes.ClusterSyncObjects{}, err
	} else if len(value) == 0 {
		return mtypes.ClusterSyncObjects{}, pkgerrors.New("Cluster sync object not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := mtypes.ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return mtypes.ClusterSyncObjects{}, err
		}
		return ckvp, nil
	}

	return mtypes.ClusterSyncObjects{}, pkgerrors.New("Unknown Error")
}

// DeleteClusterSyncObjects the  ClusterSyncObjects from database
func (c *CloudConfigClient) DeleteClusterSyncObjects(provider, syncobject string) error {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	err := db.DBconn.Remove(c.db.storeName, key)
	return err
}

// GetClusterSyncObjectsValue returns the value of the key from the corresponding provider and Sync Object name
func (c *CloudConfigClient) GetClusterSyncObjectsValue(provider, syncobject, syncobjectkey string) (interface{}, error) {
	//Construct key and tag to select entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: syncobject,
	}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return mtypes.ClusterSyncObjects{}, err
	} else if len(value) == 0 {
		return mtypes.ClusterSyncObjects{}, pkgerrors.New("Cluster sync object not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := mtypes.ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return nil, err
		}

		for _, kvmap := range ckvp.Spec.Kv {
			if val, ok := kvmap[syncobjectkey]; ok {
				return struct {
					Value interface{} `json:"value"`
				}{Value: val}, nil
			}
		}
		return nil, pkgerrors.New("Cluster Sync Object key value not found")
	}

	return nil, pkgerrors.New("Unknown Error")
}

// GetAllClusterSyncObjects returns the Cluster Sync Objects for corresponding provider
func (c *CloudConfigClient) GetAllClusterSyncObjects(provider string) ([]mtypes.ClusterSyncObjects, error) {
	//Construct key and tag to select the entry
	key := ClusterSyncObjectsKey{
		ClusterProviderName:    provider,
		ClusterSyncObjectsName: "",
	}
	values, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		return []mtypes.ClusterSyncObjects{}, err
	}

	resp := make([]mtypes.ClusterSyncObjects, 0)
	for _, value := range values {
		cp := mtypes.ClusterSyncObjects{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []mtypes.ClusterSyncObjects{}, err
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// CreateGitOpsConfig allows to create a new cloud config entry to hold a kubeconfig for access
func (c *CloudConfigClient) CreateGitOpsConfig(provider string, cluster string, gs mtypes.GitOpsSpec, level string, namespace string) (CloudGitOpsConfig, error) {

	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	// check if it already exists
	_, err := c.GetGitOpsConfig(provider, cluster, level, namespace)
	if err == nil {
		log.Error("CloudConfig already exists", log.Fields{})
		return CloudGitOpsConfig{}, pkgerrors.New("CloudConfig already exists")
	}
	log.Info("Inserting in gs db", log.Fields{"gs": gs})
	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagMeta, gs)
	if err != nil {
		log.Error("Failure inserting CloudConfig", log.Fields{})
		return CloudGitOpsConfig{}, pkgerrors.Wrap(err, "Failure inserting CloudConfig")
	}

	cc := CloudGitOpsConfig{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		Config:    gs,
	}

	return cc, nil
}

// GetGitOpsConfig allows to create a new cloud config entry to hold a kubeconfig for access
func (c *CloudConfigClient) GetGitOpsConfig(provider string, cluster string, level string, namespace string) (CloudGitOpsConfig, error) {

	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	value, err := db.DBconn.Find(c.db.storeName, key, c.db.tagMeta)
	if err != nil {
		log.Error("Failure inserting CloudConfig", log.Fields{})
		return CloudGitOpsConfig{}, pkgerrors.Wrap(err, "Failure inserting CloudConfig")
	}
	log.Info("Get in gs db", log.Fields{"value": value})

	cp := mtypes.GitOpsSpec{}
	err = db.DBconn.Unmarshal(value[0], &cp)
	if err != nil {
		return CloudGitOpsConfig{}, err
	}

	cc := CloudGitOpsConfig{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		Config:    cp,
	}

	return cc, nil
}

// DeleteGitOpsConfig deletes a cloud config entry
func (c *CloudConfigClient) DeleteGitOpsConfig(provider string, cluster string, level string, namespace string) error {
	key := CloudConfigKey{
		Provider:  provider,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
	}

	// check if it doesn't exist
	_, err := c.GetGitOpsConfig(provider, cluster, level, namespace)
	if err != nil {
		log.Error("Could not fetch the CloudConfig so not deleting", log.Fields{})
		return pkgerrors.New("Could not fetch the GitOpsConfig so not deleting")
	}

	err = db.DBconn.Remove(c.db.storeName, key)
	if err != nil {
		log.Error("Could not delete the GitOpsConfig", log.Fields{})
		return pkgerrors.Wrap(err, "Could not delete the GitOpsConfig")
	}

	return nil
}
