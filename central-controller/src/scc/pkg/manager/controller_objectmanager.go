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
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"io"
)

// ControllerManager is an interface exposes the ControllerObject functionality
type ControllerObjectManager interface {
	GetStoreName() string
	GetStoreMeta() string
	GetDepResManagers() []ControllerObjectManager
	AddDepResManager(mgr ControllerObjectManager)
	GetOwnResManagers() []ControllerObjectManager
	AddOwnResManager(mgr ControllerObjectManager)

	GetResourceName() string
	IsOperationSupported(oper string) bool
	GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error)
	CreateEmptyObject() module.ControllerObject
	ParseObject(r io.Reader) (module.ControllerObject, error)
	CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error)
	GetObject(m map[string]string) (module.ControllerObject, error)
	GetObjects(m map[string]string) ([]module.ControllerObject, error)
	UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error)
	DeleteObject(m map[string]string) error
}

type BaseObjectManager struct {
	storeName      string
	tagMeta        string
	depResManagers []ControllerObjectManager
	ownResManagers []ControllerObjectManager
}

func (c *BaseObjectManager) GetStoreName() string {
	return c.storeName
}

func (c *BaseObjectManager) GetStoreMeta() string {
	return c.tagMeta
}

func (c *BaseObjectManager) GetDepResManagers() []ControllerObjectManager {
	return c.depResManagers
}

func (c *BaseObjectManager) AddDepResManager(mgr ControllerObjectManager) {
	c.depResManagers = append(c.depResManagers, mgr)
}

func (c *BaseObjectManager) GetOwnResManagers() []ControllerObjectManager {
	return c.ownResManagers
}

func (c *BaseObjectManager) AddOwnResManager(mgr ControllerObjectManager) {
	c.ownResManagers = append(c.ownResManagers, mgr)
}
