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
	"io"

	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

type ResourceObjectKey struct {
	Cluster string `json:"cluster-info"`
	Type	string `json:"type"`
	Name	string `json:"name"`
}

// ResourceObjectManager implements the ControllerObjectManager
type ResourceObjectManager struct {
	BaseObjectManager
}

func NewResourceObjectManager() *ResourceObjectManager {
	return &ResourceObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "resource",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *ResourceObjectManager) GetResourceName() string {
	return Resource
}

func (c *ResourceObjectManager) IsOperationSupported(oper string) bool {
	return false
}

func (c *ResourceObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.ResourceObject{}
}

func (c *ResourceObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	// Currently no collections fetching supported
	return ResourceObjectKey{
		Cluster: m[OverlayResource] + "-" + m[DeviceResource],
		Type:  m["Type"],
		Name:  m["Name"],
	}, nil
}

func (c *ResourceObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.ResourceObject
	err := json.NewDecoder(r).Decode(&v)
	return &v, err
}

func (c *ResourceObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().CreateObject(c, m, t)
	return t, err
}

func (c *ResourceObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)
	return t, err
}

func (c *ResourceObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	return []module.ControllerObject{}, pkgerrors.New("Not implemented")
}

func (c *ResourceObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)
	return t, err
}

func (c *ResourceObjectManager) DeleteObject(m map[string]string) error {
	// DB Operation
	err := GetDBUtils().DeleteObject(c, m)
	return err
}
