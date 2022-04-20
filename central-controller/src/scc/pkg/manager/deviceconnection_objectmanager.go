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
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"io"
	"strings"
)

type DeviceConnObjectKey struct {
	OverlayName string `json:"overlay-name"`
	DeviceName  string `json:"device-name"`
	ConnName    string `json:"connection-name"`
}

// DeviceConnObjectManager implements the ControllerObjectManager
type DeviceConnObjectManager struct {
	BaseObjectManager
}

func NewDeviceConnObjectManager() *DeviceConnObjectManager {
	return &DeviceConnObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "deviceconn",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *DeviceConnObjectManager) GetResourceName() string {
	return ConnectionResource
}

func (c *DeviceConnObjectManager) IsOperationSupported(oper string) bool {
	if oper == "GETS" {
		return true
	}
	return false
}

func (c *DeviceConnObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.ConnectionObject{}
}

func (c *DeviceConnObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	device_name := m[DeviceResource]
	key := DeviceConnObjectKey{
		OverlayName: overlay_name,
		DeviceName:  device_name,
		ConnName:    "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.ConnectionObject)
	meta_name := to.Metadata.Name
	res_name := m[ConnectionResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.ConnName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.ConnName = meta_name
	}

	return key, nil
}

func (c *DeviceConnObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.ConnectionObject
	err := json.NewDecoder(r).Decode(&v)

	return &v, err
}

func (c *DeviceConnObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *DeviceConnObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *DeviceConnObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	overlay_name := m[OverlayResource]
	device_name := m[DeviceResource]

	return GetConnectionManager().GetObjects(overlay_name, module.CreateEndName("Device", device_name))
}

func (c *DeviceConnObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *DeviceConnObjectManager) DeleteObject(m map[string]string) error {
	return pkgerrors.New("Not implemented")
}

func (c *DeviceConnObjectManager) GetConnectedHubs(overlay_name string, device_name string) ([]string, error) {
	m := make(map[string]string)
	m[OverlayResource] = overlay_name
	m[DeviceResource] = device_name

	// get all connections
	cs, err := c.GetObjects(m)
	if err != nil {
		return []string{}, err
	}

	var hub_names []string
	for _, c := range cs {
		co := c.(*module.ConnectionObject)
		// get peer end's type and name
		t, n, ip := co.GetPeer("Device", device_name)
		if t == "Hub" {
			hub_names = append(hub_names, n+".."+ip)
		}
	}
	return hub_names, nil
}

func (c *DeviceConnObjectManager) IsConnectedHub(overlay_name string, device_name string, hub string) bool {
	hub_names, _ := c.GetConnectedHubs(overlay_name, device_name)
	for _, hub_name := range hub_names {
		strs := strings.SplitN(hub_name, "..", 2)
		if len(strs) == 2 {
			if hub == strings.Replace(strs[0], "Hub.", "", 1) {
				return true
			}
		}
	}

	return false
}
