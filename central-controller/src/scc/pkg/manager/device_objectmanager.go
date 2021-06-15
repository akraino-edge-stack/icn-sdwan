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
	"strings"
	"time"

	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	//"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/client"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	pkgerrors "github.com/pkg/errors"
)

const SCC_RESOURCE = "scc_ipsec_resource"
const RegStatus = "RegStatus"

var ips []string
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
		overlay_name := m[OverlayResource]
		oip, err := ipr_manager.Allocate("", to.Metadata.Name)
		if err != nil {
			return pkgerrors.Wrap(err, "Fail to allocate overlay ip for "+to.Metadata.Name)
		}

		// Set OIP in Device
		log.Println("Using overlay ip " + oip)
		to.Status.Ip = oip

		// Get all proposal resources
		proposal := GetManagerset().Proposal
		proposals, err := proposal.GetObjects(m)
		if len(proposals) == 0 || err != nil {
			log.Println("Missing Proposal in the overlay\n")
			return pkgerrors.New("Error in getting proposals")
		}

		var all_proposal []string
		var proposalresource []*resource.ProposalResource
		for i := 0; i < len(proposals); i++ {
			proposal_obj := proposals[i].(*module.ProposalObject)
			all_proposal = append(all_proposal, proposal_obj.Metadata.Name)
			pr := proposal_obj.ToResource()
			proposalresource = append(proposalresource, pr)
		}

		//Extract SCC cert/key
		cu, err := GetCertUtil()
		if err != nil {
			log.Println("Getting certutil error")
		}
		crts, key, err := cu.GetKeypair(SCCCertName, NameSpaceName)
		crt := strings.SplitAfter(crts, "-----END CERTIFICATE-----")[0]

		root_ca := GetRootCA(overlay_name)

		// Build up ipsec resource
		scc_conn := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(to.Metadata.Name, ""),
			ConnectionType: CONN_TYPE,
			Mode:           MODE,
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
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + SCCCertName,
			RemoteIdentifier:     "CN=" + to.GetCertName(),
			CryptoProposal:       all_proposal,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          scc_conn,
		}

		scc := module.EmptyObject{
			Metadata: module.ObjectMetaData{"local", "", "", ""}}

		// Add and deploy resource
		resutil := NewResUtil()
		resutil.AddResource(&scc, "create", &scc_ipsec_resource)
		for i := 0; i < len(proposalresource); i++ {
			resutil.AddResource(&scc, "create", proposalresource[i])
		}

		resutil.Deploy("localto"+to.Metadata.Name, "YAML")

		//Reserve ipsec resource to device object
		res_str, err := resource.GetResourceBuilder().ToString(&scc_ipsec_resource)
		to.Status.Data[SCC_RESOURCE] = res_str

		ips = append(ips, oip)

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
		for to.Status.Data[RegStatus] != "success" {
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

	device_name := m[DeviceResource]

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
		resutils.Undeploy("localto"+device_name, "YAML")
	}

	log.Println("Delete device...")
	err = overlay_manager.DeleteConnections(m, t)
	if err != nil {
		log.Println(err)
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)

	return err
}

func GetDeviceCertificate(overlay_name string, device_name string) (string, string, error) {
	cert := GetManagerset().Cert
	_, crts, key, err := cert.GetOrCreateDC(overlay_name, device_name)
	if err != nil {
		log.Println("Error in getting cert for device ...")
		return "", "", err
	}

	crt := strings.SplitAfter(crts, "-----END CERTIFICATE-----")[0]
	return crt, key, nil
}

func (c *DeviceObjectManager) PostRegister(m map[string]string, t module.ControllerObject) error {

	overlay_manager := GetManagerset().Overlay

	to := t.(*module.DeviceObject)
	log.Println("Registering device " + to.Metadata.Name + " ... ")

	if to.Status.Mode == 2 {
		kube_config, err := base64.StdEncoding.DecodeString(to.Specification.KubeConfig)
		if err != nil {
			to.Status.Data[RegStatus] = "failed"
		}

		kube_config, _, err = kubeutil.checkKubeConfigAvail(kube_config, ips, DEFAULT_K8S_API_SERVER_PORT)
		if err != nil {
			//TODO: check the error type, and if is unauthorized then switch the status to failed.
			return err
		}

		to.Status.Data[RegStatus] = "success"
		to.Specification.KubeConfig = base64.StdEncoding.EncodeToString(kube_config)
		err = GetDBUtils().RegisterDevice(to.Metadata.Name, to.Specification.KubeConfig)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("scc connection is verified.")

	} else {
		to.Status.Data[RegStatus] = "success"
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
			if to.Status.Mode == 1 || dev.Status.Mode == 1 {
				err = overlay_manager.SetupConnection(m, to, dev, DEVICETODEVICE, NameSpaceName)
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
