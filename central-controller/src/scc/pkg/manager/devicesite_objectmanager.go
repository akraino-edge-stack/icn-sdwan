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
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"io"
	"log"
)

type DeviceSiteObjectKey struct {
	OverlayName string `json:"overlay-name"`
	DeviceName  string `json:"device-name"`
	SiteName    string `json:"site-name"`
}

// DeviceSiteObjectManager implements the ControllerObjectManager
type DeviceSiteObjectManager struct {
	BaseObjectManager
}

func NewDeviceSiteObjectManager() *DeviceSiteObjectManager {
	return &DeviceSiteObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "devicesite",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *DeviceSiteObjectManager) GetResourceName() string {
	return SiteResource
}

func (c *DeviceSiteObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *DeviceSiteObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.SiteObject{}
}

func (c *DeviceSiteObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	device_name := m[DeviceResource]
	key := DeviceSiteObjectKey{
		OverlayName: overlay_name,
		DeviceName:  device_name,
		SiteName:    "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.SiteObject)
	meta_name := to.Metadata.Name
	res_name := m[SiteResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.SiteName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.SiteName = meta_name
	}

	return key, nil
}

func (c *DeviceSiteObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.SiteObject
	err := json.NewDecoder(r).Decode(&v)

	return &v, err
}

func (c *DeviceSiteObjectManager) deployResource(m map[string]string, t module.ControllerObject, update bool) error {
	overlay_name := m[OverlayResource]
	device_name := m[DeviceResource]

	to := t.(*module.SiteObject)
	site_name := to.Metadata.Name

	if len(to.Specification.Hubs) < 1 {
		return pkgerrors.New("Hub is required")
	}

	// Todo: support multiple hubs
	hub := to.Specification.Hubs[0]
	devConn := GetManagerset().DeviceConn
	if !devConn.IsConnectedHub(overlay_name, device_name, hub) {
		return pkgerrors.New("Hub does not connect to the device")
	}

	dev_manager := GetManagerset().Device
	dev, err := dev_manager.GetObject(m)
	if err != nil {
		return pkgerrors.Wrap(err, "Device "+device_name+" is not defined")
	}

	// Deploy HubSite resource to device
	devobj := dev.(*module.DeviceObject)
	resutil := NewResUtil()
	resutil.AddResource(dev, "create", &resource.HubSiteResource{
		Name:      format_resource_name(site_name, ""),
		Type:      "Device",
		Site:      to.Specification.Url,
		Subnet:    to.Specification.Subnet,
		HubIP:     "",
		DevicePIP: devobj.Status.DataIps[module.CreateEndName("Hub", hub)], // Todo: check if device mode = 1
	})

	err = resutil.DeployUpdate(overlay_name, "hubsite"+to.Metadata.Name, "YAML", update)

	return err
}

func (c *DeviceSiteObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	err := c.deployResource(m, t, false)
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)

	return t, err
}

func (c *DeviceSiteObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *DeviceSiteObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *DeviceSiteObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	err := c.deployResource(m, t, true)
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	// DB Operation
	t, err = GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *DeviceSiteObjectManager) DeleteObject(m map[string]string) error {
	overlay_name := m[OverlayResource]
	t, err := c.GetObject(m)
	if err != nil {
		log.Println(err)
		return nil
	}
	to := t.(*module.SiteObject)
	site_name := to.Metadata.Name

	dev_manager := GetManagerset().Device
	dev, err := dev_manager.GetObject(m)

	// Undeploy HubSite resource to device
	resutil := NewResUtil()
	resutil.AddResource(dev, "delete", &resource.EmptyResource{format_resource_name(site_name, ""), "HubSite"})
	err = resutil.Undeploy(overlay_name)
	if err != nil {
		log.Println(err)
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)

	return err
}

func (c *DeviceSiteObjectManager) TryDeleteSite(overlay string, hub string, device string) error {
	m := make(map[string]string)
	m[OverlayResource] = overlay
	m[DeviceResource] = device

	// get all sites and delete site which uses hub as proxy
	ts, _ := c.GetObjects(m)
	for _, t := range ts {
		to := t.(*module.SiteObject)
		if len(to.Specification.Hubs) > 0 && hub == to.Specification.Hubs[0] {
			m[SiteResource] = to.Metadata.Name
			c.DeleteObject(m)
		}
	}

	return nil
}
