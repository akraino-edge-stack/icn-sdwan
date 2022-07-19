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

	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
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
	certManager := GetManagerset().Cert
	overlay_name := m[OverlayResource]
	to := t.(*module.HubObject)
	hub_name := to.Metadata.Name

	//Todo: Check if public ip can be used.
	var local_public_ip string

	if to.Specification.KubeConfig == "" {
		var gitOpsParams mtypes.GitOpsProps
		gitOpsParams.GitOpsType = to.Specification.GitOpsParam.GitOpsType
		gitOpsParams.GitOpsReferenceObject = to.Specification.GitOpsParam.GitOpsReferenceObject
		gitOpsParams.GitOpsResourceObject = to.Specification.GitOpsParam.GitOpsResourceObject

		m[ClusterSyncResource] = gitOpsParams.GitOpsReferenceObject
		clustersync_manager := GetManagerset().ClusterSync
		clustersync_obj, err := clustersync_manager.GetObject(m)
		if clustersync_obj.GetMetadata().Name != gitOpsParams.GitOpsReferenceObject || err != nil {
			log.Println("error finding clustersync object", err)
			return &module.HubObject{}, err
		}

		if gitOpsParams.GitOpsResourceObject != "" {
			m[ClusterSyncResource] = gitOpsParams.GitOpsResourceObject
			clustersync_obj, err := clustersync_manager.GetObject(m)
			if clustersync_obj.GetMetadata().Name != gitOpsParams.GitOpsResourceObject || err != nil {
				log.Println("error finding clustersync object", err)
				return &module.HubObject{}, err
			}
		}

		to.Status.Ip = to.Specification.PublicIps[0]
		err = GetDBUtils().RegisterGitOpsDevice(overlay_name, to.Metadata.Name, mtypes.GitOpsSpec{Props: gitOpsParams})
		if err != nil {
			log.Println(err)
		}
	} else {
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
			err := GetDBUtils().RegisterDevice(overlay_name, hub_name, to.Specification.KubeConfig)
			if err != nil {
				log.Println(err)
			}
		} else {
			return t, err
		}
	}

	//Create cert for ipsec connection
	log.Println("Create Certificate: " + to.GetCertName())
	_, _, _, err := certManager.GetOrCreateCertificateByType(overlay_name, to.Metadata.Name, HubKey, false)
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
	certManager := GetManagerset().Cert

	// Reset all IpSec connection setup by this device
	err = overlay_manager.DeleteConnections(m, t)
	if err != nil {
		log.Println(err)
	}

	to := t.(*module.HubObject)
	log.Println("Delete Certificate: " + to.GetCertName())
	err = certManager.DeleteCertificateByType(m[OverlayResource], to.Metadata.Name, HubKey)
	if err != nil {
		log.Println("Error in deleting hub certificate")
	}
	//overlay_manager.DeleteCertificate(to.GetCertName())

	if to.Specification.KubeConfig == "" {
		err = GetDBUtils().UnregisterGitOpsDevice(m[OverlayResource], to.Metadata.Name)
	} else {
		err = GetDBUtils().UnregisterDevice(m[OverlayResource], m[HubResource])
	}
	if err != nil {
		log.Println(err)
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)
	if err != nil {
		log.Println(err)
	}

	return err
}
