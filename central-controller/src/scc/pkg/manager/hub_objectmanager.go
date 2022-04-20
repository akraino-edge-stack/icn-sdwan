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
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
)

const DEFAULTPORT = "6443"

type HubObjectKey struct {
	OverlayName string `json:"overlay-name"`
	HubName     string `json:"hub-name"`
}

// HubObjectManager implements the ControllerObjectManager
type HubObjectManager struct {
	BaseObjectManager
}

func NewHubObjectManager() *HubObjectManager {
	return &HubObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "hub",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *HubObjectManager) GetResourceName() string {
	return HubResource
}

func (c *HubObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *HubObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.HubObject{}
}

func (c *HubObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	key := HubObjectKey{
		OverlayName: overlay_name,
		HubName:     "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.HubObject)
	meta_name := to.Metadata.Name
	res_name := m[HubResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.HubName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.HubName = meta_name
	}

	return key, nil
}

func (c *HubObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.HubObject
	err := json.NewDecoder(r).Decode(&v)

	// initial Status
	v.Status.Data = make(map[string]string)
	return &v, err
}

func (c *HubObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	overlay := GetManagerset().Overlay
	overlay_name := m[OverlayResource]
	to := t.(*module.HubObject)
	hub_name := to.Metadata.Name

	//Todo: Check if public ip can be used.
	var local_public_ip string
	var config []byte
	config, err := base64.StdEncoding.DecodeString(to.Specification.KubeConfig)
	if err != nil {
		log.Println(err)
		return t, err
	}

	local_public_ips := to.Specification.PublicIps

	kubeutil := GetKubeConfigUtil()
	config, local_public_ip, err = kubeutil.checkKubeConfigAvail(config, local_public_ips, DEFAULTPORT)
	if err == nil {
		log.Println("Public IP address verified: " + local_public_ip)
		to.Status.Ip = local_public_ip
		to.Specification.KubeConfig = base64.StdEncoding.EncodeToString(config)
		err := GetDBUtils().RegisterDevice(hub_name, to.Specification.KubeConfig)
		if err != nil {
			log.Println(err)
		}
	} else {
		return t, err
	}

	//Create cert for ipsec connection
	log.Println("Create Certificate: " + to.GetCertName())
	_, _, err = overlay.CreateCertificate(overlay_name, to.GetCertName())
	if err != nil {
		return t, err
	}

	//Get all available hub objects
	hubs, err := c.GetObjects(m)
	if err != nil {
		log.Println(err)
	}

	//TODO: Need to add funcs to re-create connections if some of the connections are not ready
	//Maybe because of cert not ready or other reasons.
	if len(hubs) > 0 && err == nil {
		for i := 0; i < len(hubs); i++ {
			err := overlay.SetupConnection(m, t, hubs[i], HUBTOHUB, NameSpaceName, false)
			if err != nil {
				log.Println("Setup connection with " + hubs[i].(*module.HubObject).Metadata.Name + " failed.")
			}
		}
		t, err = GetDBUtils().CreateObject(c, m, t)
	} else {

		t, err = GetDBUtils().CreateObject(c, m, t)
	}

	return t, err
}

func (c *HubObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *HubObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *HubObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *HubObjectManager) DeleteObject(m map[string]string) error {
	//Check resource exists

	t, err := c.GetObject(m)
	if err != nil {
		return nil
	}

	overlay_manager := GetManagerset().Overlay

	// Reset all IpSec connection setup by this device
	err = overlay_manager.DeleteConnections(m, t)
	if err != nil {
		log.Println(err)
	}

	to := t.(*module.HubObject)
	log.Println("Delete Certificate: " + to.GetCertName())
	overlay_manager.DeleteCertificate(to.GetCertName())

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)
	err = GetDBUtils().UnregisterDevice(m[HubResource])
	if err != nil {
		log.Println(err)
	}

	return err
}

func GetHubCertificate(cert_name string, namespace string) (string, string, error) {
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
		return "", "", err
	} else {
		ready := cu.IsCertReady(cert_name, namespace)
		if ready != true {
			return "", "", pkgerrors.New("Cert for hub is not ready")
		} else {
			crts, key, err := cu.GetKeypair(cert_name, namespace)
			crt := strings.SplitAfter(crts, "-----END CERTIFICATE-----")[0]
			if err != nil {
				log.Println(err)
				return "", "", err
			}
			return crt, key, nil
		}
	}
}
