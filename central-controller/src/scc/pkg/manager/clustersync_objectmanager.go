/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package manager

import (
	"encoding/json"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	pkgerrors "github.com/pkg/errors"
	"io"
)

type ClusterSyncObjectKey struct {
	OverlayName     string `json:"overlay-name"`
	ClusterSyncName string `json:"clustersync-name"`
}

// ClusterSyncObjectManager implements the ControllerObjectManager
type ClusterSyncObjectManager struct {
	BaseObjectManager
}

func NewClusterSyncObjectManager() *ClusterSyncObjectManager {
	return &ClusterSyncObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "clustersync",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *ClusterSyncObjectManager) GetResourceName() string {
	return ClusterSyncResource
}

func (c *ClusterSyncObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *ClusterSyncObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.ClusterSyncObject{}
}

func (c *ClusterSyncObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	key := ClusterSyncObjectKey{
		OverlayName:     overlay_name,
		ClusterSyncName: "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.ClusterSyncObject)
	meta_name := to.Metadata.Name
	res_name := m[ClusterSyncResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.ClusterSyncName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.ClusterSyncName = meta_name
	}

	return key, nil
}

func (c *ClusterSyncObjectManager) getProvider(overlay string) string {
	return PROVIDERNAME + "_" + overlay
}

func (c *ClusterSyncObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.ClusterSyncObject
	err := json.NewDecoder(r).Decode(&v)

	return &v, err
}

func (c *ClusterSyncObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	ccc := rsync.NewCloudConfigClient()
	to := t.(*module.ClusterSyncObject)

	overlay_name := m[OverlayResource]
	_, err := ccc.CreateClusterSyncObjects(c.getProvider(overlay_name), 
		mtypes.ClusterSyncObjects{
			Metadata: mtypes.Metadata{
				Name: to.Metadata.Name,
				Description: to.Metadata.Description,
				UserData1: to.Metadata.UserData1,
				UserData2: to.Metadata.UserData2,
			},
			Spec: mtypes.ClusterSyncObjectSpec{
				Kv: to.Specification.Kv,
			},
		}, false)

	if err != nil {
		return c.CreateEmptyObject(), err
	}

	return t, nil
}

func (c *ClusterSyncObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	ccc := rsync.NewCloudConfigClient()
	overlay_name := m[OverlayResource]
	res_name := m[ClusterSyncResource]

	so, err := ccc.GetClusterSyncObjects(c.getProvider(overlay_name), res_name)
	if err != nil {
		return c.CreateEmptyObject(), err
	}

	return &module.ClusterSyncObject{
		Metadata: module.ObjectMetaData{
			Name: so.Metadata.Name,
			Description: so.Metadata.Description,
			UserData1: so.Metadata.UserData1,
			UserData2: so.Metadata.UserData2,
		},
		Specification: module.ClusterSyncObjectSpec{
			Kv: so.Spec.Kv,
		},
	}, nil
}

func (c *ClusterSyncObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	ccc := rsync.NewCloudConfigClient()
	overlay_name := m[OverlayResource]

	sos, err := ccc.GetAllClusterSyncObjects(c.getProvider(overlay_name))
	if err != nil {
		return []module.ControllerObject{}, err
	}

	var resp []module.ControllerObject
	for _, so := range sos {
		resp = append(resp, 
			&module.ClusterSyncObject{
				Metadata: module.ObjectMetaData{
					Name: so.Metadata.Name,
					Description: so.Metadata.Description,
					UserData1: so.Metadata.UserData1,
					UserData2: so.Metadata.UserData2,
				},
				Specification: module.ClusterSyncObjectSpec{
					Kv: so.Spec.Kv,
				},
			})
	}

	return resp, nil
}

func (c *ClusterSyncObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	ccc := rsync.NewCloudConfigClient()
	to := t.(*module.ClusterSyncObject)

	overlay_name := m[OverlayResource]
	_, err := ccc.CreateClusterSyncObjects(c.getProvider(overlay_name), 
		mtypes.ClusterSyncObjects{
			Metadata: mtypes.Metadata{
				Name: to.Metadata.Name,
				Description: to.Metadata.Description,
				UserData1: to.Metadata.UserData1,
				UserData2: to.Metadata.UserData2,
			},
			Spec: mtypes.ClusterSyncObjectSpec{
				Kv: to.Specification.Kv,
			},
		}, true)

	if err != nil {
		return c.CreateEmptyObject(), err
	}

	return t, nil
}

func (c *ClusterSyncObjectManager) DeleteObject(m map[string]string) error {
	ccc := rsync.NewCloudConfigClient()
	overlay_name := m[OverlayResource]
	res_name := m[ClusterSyncResource]

	return ccc.DeleteClusterSyncObjects(c.getProvider(overlay_name), res_name)
}