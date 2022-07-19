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
	"io"
	"log"
	//"strconv"
	"encoding/base64"
	"encoding/json"
	"github.com/matryer/runner"
	"time"

	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	//"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/client"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	pkgerrors "github.com/pkg/errors"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

const SCC_RESOURCE = "scc_ipsec_resource"
const RegStatus = "RegStatus"
const globalOverlay = "global"

var task *runner.Task

type DeviceObjectKey struct {
	OverlayName string `json:"overlay-name"`
	DeviceName  string `json:"device-name"`
}

// DeviceObjectManager implements the ControllerObjectManager
type DeviceObjectManager struct {
	BaseObjectManager
}

func NewDeviceObjectManager() *DeviceObjectManager {
	return &DeviceObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "device",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *DeviceObjectManager) GetResourceName() string {
	return DeviceResource
}

func (c *DeviceObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *DeviceObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.DeviceObject{}
}

func (c *DeviceObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]
	key := DeviceObjectKey{
		OverlayName: overlay_name,
		DeviceName:  "",
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.DeviceObject)
	meta_name := to.Metadata.Name
	res_name := m[DeviceResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.DeviceName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.DeviceName = meta_name
	}

	return key, nil
}

func (c *DeviceObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.DeviceObject
	err := json.NewDecoder(r).Decode(&v)

	// initial Status
	v.Status.Data = make(map[string]string)
	v.Status.DataIps = make(map[string]string)
	return &v, err
}

func (c *DeviceObjectManager) PreProcessing(m map[string]string, t module.ControllerObject) error {
	to := t.(*module.DeviceObject)

	ipr_manager := GetManagerset().ProviderIPRange

	if to.Specification.KubeConfig == "" {
		to.Status.Mode = 3
		to.Status.Data[RegStatus] = "pending"

		gitOpsReference := to.Specification.GitOpsParam.GitOpsReferenceObject
		gitOpsResource := to.Specification.GitOpsParam.GitOpsResourceObject

		m[ClusterSyncResource] = gitOpsReference
		clustersync_manager := GetManagerset().ClusterSync
		clustersync_obj, err := clustersync_manager.GetObject(m)
		if clustersync_obj.GetMetadata().Name != gitOpsReference || err != nil {
			log.Println(err)
			return err
		}

		if gitOpsResource != "" {
			m[ClusterSyncResource] = gitOpsResource
			clustersync_obj, err := clustersync_manager.GetObject(m)
			if clustersync_obj.GetMetadata().Name != gitOpsResource || err != nil {
				log.Println(err)
				return err
			}
		}

		return nil
	}

	kubeutil := GetKubeConfigUtil()

	local_public_ips := to.Specification.PublicIps
	kube_config, err := base64.StdEncoding.DecodeString(to.Specification.KubeConfig)
	if err != nil {
		return pkgerrors.Wrap(err, "Fail to decode kubeconfig")
	}

	// Set the Register status to pending
	to.Status.Data[RegStatus] = "pending"

	if len(local_public_ips) > 0 {
		// Use public IP as external connection
		to.Status.Mode = 1

		kube_config, local_public_ip, err := kubeutil.checkKubeConfigAvail(kube_config, local_public_ips, "6443")
		if err != nil {
			return pkgerrors.Wrap(err, "Fail to verify public ip")
		}

		// Set IP in device
		log.Println("Use public ip " + local_public_ip)
		to.Status.Ip = local_public_ip

		// Set new kubeconfig in device
		to.Specification.KubeConfig = base64.StdEncoding.EncodeToString([]byte(kube_config))
	} else {
		// Use scc as external connection
		to.Status.Mode = 2

		// allocate OIP for device
		oip, err := ipr_manager.Allocate("", to.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrap(err, "Fail to allocate overlay ip for "+to.Metadata.Name)
		}

		// Set OIP in Device
		log.Println("Using overlay ip " + oip)
		to.Status.Ip = oip

		resutil := NewResUtil()
		scc := module.EmptyObject{
			Metadata: module.ObjectMetaData{"local", "", "", ""}}

		// Get all proposal resources
		proposal := GetManagerset().Proposal
		proposals, err := proposal.GetObjects(m)
		if len(proposals) == 0 || err != nil {
			log.Println("Missing Proposal in the overlay\n")
			return pkgerrors.New("Error in getting proposals")
		}

		var all_proposal []string
		for i := 0; i < len(proposals); i++ {
			proposal_obj := proposals[i].(*module.ProposalObject)
			all_proposal = append(all_proposal, proposal_obj.Metadata.Name)
			pr := proposal_obj.ToResource()
			resutil.AddResource(&scc, "create", pr)
		}

		//Extract SCC cert/key
		cu, err := GetCertUtil()
		if err != nil {
			log.Println("Getting certutil error")
		}

		crt, key, err := cu.GetKeypair(SCCCertName, NameSpaceName)
		root_ca, _ := GetRootCA()

		// Build up ipsec resource
		scc_conn := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(to.Metadata.Name, ""),
			ConnectionType: CONN_TYPE,
			Mode:           START_MODE,
			Mark:           DEFAULT_MARK,
			RemoteSourceIp: oip,
			LocalUpDown:    DEFAULT_UPDOWN,
			CryptoProposal: all_proposal,
		}

		scc_ipsec_resource := resource.IpsecResource{
			Name:                 "localto" + format_resource_name(to.Metadata.Name, ""),
			Type:                 VTI_MODE,
			Remote:               ANY,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           crt,
			PrivateCert:          key,
			SharedCA:             root_ca,
			LocalIdentifier:      "CN=" + SCCCertName,
			RemoteIdentifier:     "CN=" + to.GetCertName(),
			CryptoProposal:       all_proposal,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          scc_conn,
		}

		// Add and deploy resource
		resutil.AddResource(&scc, "create", &scc_ipsec_resource)
		resutil.Deploy(globalOverlay, m[OverlayResource]+"localto"+to.Metadata.Name, "YAML")

		//Reserve ipsec resource to device object
		res_str, err := resource.GetResourceBuilder().ToString(&scc_ipsec_resource)
		to.Status.Data[SCC_RESOURCE] = res_str

	}
	return nil

}

func (c *DeviceObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	err := c.PreProcessing(m, t)
	if err != nil {
		return c.CreateEmptyObject(), err
	}

	to := t.(*module.DeviceObject)
	task = runner.Go(func(ShouldStop runner.S) error {
		for to.Status.Data[RegStatus] == "pending" {
			err = c.PostRegister(m, t)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(5 * time.Second)
			if ShouldStop() {
				break
			}
		}
		return nil
	})

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)
	return t, err
}

func (c *DeviceObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *DeviceObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *DeviceObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *DeviceObjectManager) DeleteObject(m map[string]string) error {
	t, err := c.GetObject(m)
	if err != nil {
		return nil
	}

	if task != nil && task.Running() {
		task.Stop()
		select {
		case <-task.StopChan():
		case <-time.After(2 * time.Second):
			log.Println("Goroutine register device stopped")
		}
	}

	overlay_manager := GetManagerset().Overlay
	ipr_manager := GetManagerset().ProviderIPRange
	cert_manager := GetManagerset().Cert

	overlay_name := m[OverlayResource]

	to := t.(*module.DeviceObject)

	//If the device is in mode 2:
	// * Free OIP assigned
	// * Remove ipsec configuration on SCC
	if to.Status.Mode == 2 {
		// Free OIP
		ipr_manager.Free("", to.Status.Ip)

		scc := module.EmptyObject{
			Metadata: module.ObjectMetaData{"local", "", "", ""}}

		resutils := NewResUtil()
		r_str := to.Status.Data["scc_ipsec_resource"]
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutils.AddResource(&scc, "create", r)

		// Get all proposal resources
		proposal := GetManagerset().Proposal
		proposals, err := proposal.GetObjects(m)
		if len(proposals) == 0 || err != nil {
			log.Println("Missing Proposal in the overlay")
			return pkgerrors.New("Error in getting proposals")
		}

		for i := 0; i < len(proposals); i++ {
			proposal_obj := proposals[i].(*module.ProposalObject)
			pr := proposal_obj.ToResource()
			resutils.AddResource(&scc, "create", pr)
		}

		resutils.Undeploy(globalOverlay)
	}

	log.Println("Delete device...")
	err = overlay_manager.DeleteConnections(m, t)
	if err != nil {
		log.Println(err)
	}

	if to.Status.Mode == 3 {
		err = GetDBUtils().UnregisterGitOpsDevice(overlay_name, to.Metadata.Name)
	} else {
		err = GetDBUtils().UnregisterDevice(overlay_name, m[DeviceResource])
	}
	if err != nil {
		log.Println(err)
	}

	log.Println("Delete Certificate: " + to.GetCertName())
	err = cert_manager.DeleteCertificateByType(overlay_name, to.Metadata.Name, DeviceKey)
	if err != nil {
		log.Println("Error in deleting device certificate")
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (c *DeviceObjectManager) PostRegister(m map[string]string, t module.ControllerObject) error {
	overlay_name := m[OverlayResource]
	overlay_manager := GetManagerset().Overlay
	cert_manager := GetManagerset().Cert

	to := t.(*module.DeviceObject)
	log.Println("Registering device " + to.Metadata.Name + " ... ")

	if to.Status.Mode == 3 {
		to.Status.Data[RegStatus] = "success"
		var gitOpsParams mtypes.GitOpsProps
		gitOpsParams.GitOpsType = to.Specification.GitOpsParam.GitOpsType
		gitOpsParams.GitOpsReferenceObject = to.Specification.GitOpsParam.GitOpsReferenceObject
		gitOpsParams.GitOpsResourceObject = to.Specification.GitOpsParam.GitOpsResourceObject

		err := GetDBUtils().RegisterGitOpsDevice(overlay_name, to.Metadata.Name, mtypes.GitOpsSpec{Props: gitOpsParams})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println("Create Certificate: " + to.GetCertName())
		_, _, _, err = cert_manager.GetOrCreateDC(overlay_name, to.Metadata.Name, false)
		if err != nil {
			log.Println(err)
			return err
		}
	} else if to.Status.Mode == 2 {
		kube_config, err := base64.StdEncoding.DecodeString(to.Specification.KubeConfig)
		if err != nil || len(kube_config) == 0 {
			to.Status.Data[RegStatus] = "failed"
			return pkgerrors.New("Error in decoding kubeconfig in registration")
		}

		kube_config, _, err = kubeutil.checkKubeConfigAvail(kube_config, []string{to.Status.Ip}, DEFAULT_K8S_API_SERVER_PORT)
		if err != nil {
			//TODO: check the error type, and if is unauthorized then switch the status to failed.
			return err
		}

		to.Status.Data[RegStatus] = "success"
		to.Specification.KubeConfig = base64.StdEncoding.EncodeToString(kube_config)
		err = GetDBUtils().RegisterDevice(overlay_name, to.Metadata.Name, to.Specification.KubeConfig)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("scc connection is verified.")

	} else {
		to.Status.Data[RegStatus] = "success"
		err := GetDBUtils().RegisterDevice(overlay_name, to.Metadata.Name, to.Specification.KubeConfig)
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println("Create Certificate: " + to.GetCertName())
		_, _, _, err = cert_manager.GetOrCreateDC(overlay_name, to.Metadata.Name, false)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if to.Status.Data[RegStatus] == "success" {
		devices, err := c.GetObjects(m)
		if err != nil {
			log.Println(err)
			return err
		}

		//TODO: Need to add funcs to re-create connections if some of the connections are not ready
		//Maybe because of cert not ready or other reasons.
		for i := 0; i < len(devices); i++ {
			dev := devices[i].(*module.DeviceObject)
			if (dev.Metadata.Name != to.Metadata.Name) && (to.Status.Mode == 1 || dev.Status.Mode == 1) {
				err = overlay_manager.SetupConnection(m, to, dev, DEVICETODEVICE, NameSpaceName, false)
				if err != nil {
					return err
				}
			}
		}
	}

	c.UpdateObject(m, t)
	return nil
}

//Function allocate ip and update
func (c *DeviceObjectManager) AllocateIP(m map[string]string, t module.ControllerObject, name string) (string, error) {
	to := t.(*module.DeviceObject)
	overlay_name := m[OverlayResource]
	ipr_manager := GetManagerset().IPRange

	// Allocate OIP for the device
	oip, err := ipr_manager.Allocate(overlay_name, to.Metadata.Name)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Fail to allocate overlay ip for "+to.Metadata.Name)
	}
	// Record the OIP allocated in the 'Status'
	to.Status.DataIps[name] = oip
	log.Println("Allocate DataIp name:" + name)

	c.UpdateObject(m, t)
	return oip, nil
}

//Function free ip and update
func (c *DeviceObjectManager) FreeIP(m map[string]string, t module.ControllerObject, name string) error {
	to := t.(*module.DeviceObject)
	overlay_name := m[OverlayResource]
	ipr_manager := GetManagerset().IPRange

	log.Println(to.Status.DataIps)
	oip := to.Status.DataIps[name]
	log.Println("Free DataIp name:" + name + " with ip" + oip)

	//Free the OIP
	err := ipr_manager.Free(overlay_name, oip)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to free overlay ip for connection with"+to.Metadata.Name)
	}
	log.Println("Delete ip from dataips...")
	delete(to.Status.DataIps, name)

	c.UpdateObject(m, t)
	return nil
}
