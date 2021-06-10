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
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
	"io"
)

type ProposalObjectKey struct {
	OverlayName  string `json:"overlay-name"`
	ProposalName string `json:"proposal-name"`
}

// ProposalObjectManager implements the ControllerObjectManager
type ProposalObjectManager struct {
	BaseObjectManager
}

func NewProposalObjectManager() *ProposalObjectManager {
	return &ProposalObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "proposal",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *ProposalObjectManager) GetResourceName() string {
	return ProposalResource
}

func (c *ProposalObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *ProposalObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.ProposalObject{}
}

func (c *ProposalObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	key := ProposalObjectKey{
		OverlayName:  overlay_name,
		ProposalName: "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.ProposalObject)
	meta_name := to.Metadata.Name
	res_name := m[ProposalResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.ProposalName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.ProposalName = meta_name
	}

	return key, nil
}

func (c *ProposalObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.ProposalObject
	err := json.NewDecoder(r).Decode(&v)

	return &v, err
}

func (c *ProposalObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().CreateObject(c, m, t)

	return t, err
}

func (c *ProposalObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *ProposalObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *ProposalObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *ProposalObjectManager) DeleteObject(m map[string]string) error {
	// DB Operation
	err := GetDBUtils().DeleteObject(c, m)

	return err
}
