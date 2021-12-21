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
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
	"io"
	"log"
	"strings"
)

const (
 DEFAULT_MARK = "30"
 VTI_MODE = "VTI-based"
 POLICY_MODE = "policy-based"
 PUBKEY_AUTH = "pubkey"
 FORCECRYPTOPROPOSAL = "0"
 DEFAULT_CONN = "Conn"
 DEFAULT_UPDOWN = "/etc/updown"
 IPTABLES_UPDOWN = "/usr/lib/ipsec/_updown iptables"
 OIP_UPDOWN = "/etc/updown_oip"
 CONN_TYPE = "tunnel"
 START_MODE = "start"
 ADD_MODE = "add"
 OVERLAYIP = "overlayip"
 HUBTOHUB = "hub-to-hub"
 HUBTODEVICE = "hub-to-device"
 DEVICETODEVICE = "device-to-device"
 BYCONFIG = "%config"
 ANY = "%any"
 BASE_PROTOCOL = "TCP"
 DEFAULT_K8S_API_SERVER_PORT = "6443"
 ACCEPT = "ACCEPT"
 WILDCARD_SUBNET="0.0.0.0"
)

type OverlayObjectKey struct {
	OverlayName string `json:"overlay-name"`
}

// OverlayObjectManager implements the ControllerObjectManager
type OverlayObjectManager struct {
	BaseObjectManager
}

func NewOverlayObjectManager() *OverlayObjectManager {
	return &OverlayObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "overlay",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *OverlayObjectManager) GetResourceName() string {
	return OverlayResource
}

func (c *OverlayObjectManager) IsOperationSupported(oper string) bool {
	return true
}

func (c *OverlayObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.OverlayObject{}
}

func (c *OverlayObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	key := OverlayObjectKey{""}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.OverlayObject)
	meta_name := to.Metadata.Name
	res_name := m[OverlayResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.OverlayName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.OverlayName = meta_name
	}

	return key, nil
}

func (c *OverlayObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.OverlayObject
	err := json.NewDecoder(r).Decode(&v)

	return &v, err
}

func (c *OverlayObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// Create a issuer each overlay
	to := t.(*module.OverlayObject)
	overlay_name := to.Metadata.Name
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
	} else {
		// create overlay ca
		_, err := cu.CreateCertificate(c.CertName(overlay_name), NameSpaceName, RootCAIssuerName, true)
		if err == nil {
			// create overlay issuer
			_, err := cu.CreateCAIssuer(c.IssuerName(overlay_name), NameSpaceName, c.CertName(overlay_name))
			if err != nil {
				log.Println("Failed to create overlay[" + overlay_name + "] issuer: " + err.Error())
			}
		} else {
			log.Println("Failed to create overlay[" + overlay_name + "] certificate: " + err.Error())
		}
	}

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)

	return t, err
}

func (c *OverlayObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *OverlayObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *OverlayObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *OverlayObjectManager) DeleteObject(m map[string]string) error {
	overlay_name := m[OverlayResource]

	// DB Operation
	err := GetDBUtils().DeleteObject(c, m)
	if err == nil {
		cu, err := GetCertUtil()
		if err != nil {
			log.Println(err)
		} else {
			err = cu.DeleteIssuer(c.IssuerName(overlay_name), NameSpaceName)
			if err != nil {
				log.Println("Failed to delete overlay[" + overlay_name + "] issuer: " + err.Error())
			}
			err = cu.DeleteCertificate(c.CertName(overlay_name), NameSpaceName)
			if err != nil {
				log.Println("Failed to delete overlay[" + overlay_name + "] certificate: " + err.Error())
			}
		}
	}

	return err
}

func (c *OverlayObjectManager) IssuerName(name string) string {
	return name + "-issuer"
}

func (c *OverlayObjectManager) CertName(name string) string {
	return name + "-cert"
}

func (c *OverlayObjectManager) CreateCertificate(oname string, cname string) (string, string, error) {
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
	} else {
		_, err := cu.CreateCertificate(cname, NameSpaceName, c.IssuerName(oname), false)
		if err != nil {
			log.Println("Failed to create overlay[" + oname + "] certificate: " + err.Error())
		} else {
			crts, key, err := cu.GetKeypair(cname, NameSpaceName)
			if err != nil {
				log.Println(err)
				return "", "", err
			} else {
				crt := strings.SplitAfter(crts, "-----END CERTIFICATE-----")[0]
				return crt, key, nil
			}
		}
	}

	return "", "", nil
}

func (c *OverlayObjectManager) DeleteCertificate(cname string) (string, string, error) {
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
	} else {
		err = cu.DeleteCertificate(cname, NameSpaceName)
		if err != nil {
			log.Println("Failed to delete " + cname + " certificate: " + err.Error())
		}
	}

	return "", "", nil
}

func (c *OverlayObjectManager) GetCertificate(oname string) (string, string, error) {
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
	} else {
		cname := c.CertName(oname)
		return cu.GetKeypair(cname, NameSpaceName)
	}
	return "", "", nil
}

//Set up Connection between objects
//Passing the original map resource, the two objects, connection type("hub-to-hub", "hub-to-device", "device-to-device") and namespace name.
func (c *OverlayObjectManager) SetupConnection(m map[string]string, m1 module.ControllerObject, m2 module.ControllerObject, conntype string, namespace string, is_delegated bool) error {
	//Get all proposals available in the overlay
	resutil := NewResUtil()
	hubConn := GetManagerset().HubConn
	hub_manager := GetManagerset().Hub
	dev_manager := GetManagerset().Device
	overlay_name := m[OverlayResource]

	proposal := GetManagerset().Proposal
	proposals, err := proposal.GetObjects(m)
	if len(proposals) == 0 || err != nil {
		log.Println("Missing Proposal in the overlay\n")
		return pkgerrors.New("Error in getting proposals")
	}
	var all_proposals []string
//	var proposalresources []*resource.ProposalResource
	for i := 0; i < len(proposals); i++ {
		proposal_obj := proposals[i].(*module.ProposalObject)
		all_proposals = append(all_proposals, proposal_obj.Metadata.Name)
		pr := proposal_obj.ToResource()
//		proposalresources = append(proposalresources, pr)

		// Add proposal resources
		resutil.AddResource(m1, "create", pr)
		resutil.AddResource(m2, "create", pr)
	}

	device_mgr := GetManagerset().Device

	//Get the overlay cert
	var root_ca string
	root_ca = GetRootCA(m[OverlayResource])

	var obj1_ipsec_resource resource.IpsecResource
	var obj2_ipsec_resource resource.IpsecResource
	var obj1_ip string
	var obj2_ip string

	switch conntype {
	case HUBTOHUB:
		obj1 := m1.(*module.HubObject)
		obj2 := m2.(*module.HubObject)

		obj1_ip = obj1.Status.Ip
		obj2_ip = obj2.Status.Ip

		//Keypair
		obj1_crt, obj1_key, err := GetHubCertificate(obj1.GetCertName(), namespace)
		if err != nil {
			return err
		}
		obj2_crt, obj2_key, err := GetHubCertificate(obj2.GetCertName(), namespace)
		if err != nil {
			return err
		}

		//IpsecResources
		conn1 := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			ConnectionType: CONN_TYPE,
			Mode:           ADD_MODE,
			Mark:           DEFAULT_MARK,
			LocalSubnet:    WILDCARD_SUBNET+"/0",
			RemoteSubnet:   WILDCARD_SUBNET+"/0",
			LocalUpDown:    DEFAULT_UPDOWN,
			CryptoProposal: all_proposals,
		}
		conn2 := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			Mark:           DEFAULT_MARK,
			Mode:           START_MODE,
			ConnectionType: CONN_TYPE,
			LocalSubnet:    WILDCARD_SUBNET+"/0",
			RemoteSubnet:   WILDCARD_SUBNET+"/0",
			LocalUpDown:    DEFAULT_UPDOWN,
			CryptoProposal: all_proposals,
		}
		obj1_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			Type:                 VTI_MODE,
			Remote:               obj2_ip,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(obj1_crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(obj1_key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj1.GetCertName(),
			RemoteIdentifier:     "CN=" + obj2.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          conn1,
		}
		obj2_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj2.Metadata.Name, obj1.Metadata.Name),
			Type:                 VTI_MODE,
			Remote:               obj1_ip,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(obj2_crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(obj2_key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj2.GetCertName(),
			RemoteIdentifier:     "CN=" + obj1.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          conn2,
		}

		// for each edge connect to hub2(obj2), add Route in hub1(obj1)
		// Todo: handle the error the route rule may fail if the vti interface is not exist 
		dev_names, _ := hubConn.GetConnectedDevices(overlay_name, obj2.Metadata.Name)
		for _, dev_name := range dev_names {
			log.Println(dev_name)
			strs := strings.SplitN(dev_name, "..", 2)
			if len(strs) == 2 {
				log.Println("Route Rule in " + obj1.Metadata.Name + " : " + strs[1] + " via " + obj2.Metadata.Name)
				resutil.AddResource(m1, "create", &resource.RouteResource {
					Name: strs[1] + "-" + obj2_ip,
					Destination: strs[1],
					Device: "vti_" + obj2_ip, // Todo: use the right ifname
					Table: "default", // Todo: need check
				})
			}
		}
	case HUBTODEVICE:
		obj1 := m1.(*module.HubObject)
		obj2 := m2.(*module.DeviceObject)

		obj1_ip = obj1.Status.Ip
		obj2_ip, _ = device_mgr.AllocateIP(m, m2, module.CreateEndName(obj1.GetType(), obj1.Metadata.Name))

		//Keypair
		obj1_crt, obj1_key, err := GetHubCertificate(obj1.GetCertName(), namespace)
		if err != nil {
			return err
		}

		obj1_conn := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(obj2.Metadata.Name, ""),
			ConnectionType: CONN_TYPE,
			Mode:           START_MODE,
			Mark:           DEFAULT_MARK,
			RemoteSourceIp: obj2_ip,
			LocalUpDown:    DEFAULT_UPDOWN,
			LocalSubnet:    WILDCARD_SUBNET+"/0",
			CryptoProposal: all_proposals,
		}

		obj1_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			Type:                 VTI_MODE,
			Remote:               ANY,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(obj1_crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(obj1_key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj1.GetCertName(),
			RemoteIdentifier:     "CN=" + obj2.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          obj1_conn,
		}

		obj2_crt, obj2_key, err := GetDeviceCertificate(m[OverlayResource], obj2.Metadata.Name)
		if err != nil {
			return err
		}

		//IpsecResources
		obj2_conn := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(obj1.Metadata.Name, ""),
			Mode:           START_MODE,
			LocalUpDown:    OIP_UPDOWN,
			ConnectionType: CONN_TYPE,
			LocalSourceIp:  BYCONFIG,
			RemoteSubnet:   WILDCARD_SUBNET+"/0",
			CryptoProposal: all_proposals,
		}
		obj2_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj2.Metadata.Name, obj1.Metadata.Name),
			Type:                 POLICY_MODE,
			Remote:               obj1_ip,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           obj2_crt,
			PrivateCert:          obj2_key,
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj2.GetCertName(),
			RemoteIdentifier:     "CN=" + obj1.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          obj2_conn,
		}

		hubName := obj1.GetType() + "." + obj1.Metadata.Name

		// for each hub, add route (e.g. to obj2 via obj1)
		hubs, _ := hub_manager.GetObjects(m)
		for _, hub_obj := range hubs {
			if hub_obj.GetMetadata().Name != obj1.GetMetadata().Name {
				resutil.AddResource(hub_obj, "create", &resource.RouteResource {
					Name: obj2_ip + "-" + obj1_ip,
					Destination: obj2_ip,
					Device: "vti_" + obj1_ip, // Todo: use the right ifname
					Table: "default", // Todo: need check
				})
			}
		}
		// for each edge connect to obj1 (1) add route( e.g. to obj2 via obj1) (2) add SNAT (e.g. to obj2 --to-source edge ip) 
		dev_names, _ := hubConn.GetConnectedDevices(overlay_name, obj1.Metadata.Name)
		mm := make(map[string]string)
		mm[OverlayResource] = overlay_name

		for _, dev_name := range dev_names {
			strs := strings.SplitN(dev_name, "..", 2)
			if len(strs) == 2 {
				mm[DeviceResource] = strings.Replace(strs[0], "Device.", "", 1)
				dev_obj, err := dev_manager.GetObject(mm)
				dev := dev_obj.(*module.DeviceObject)
				if err == nil {
					log.Println("Route Rule in " + strs[0] + " : " + obj2_ip + " via " + obj1.Metadata.Name)
					resutil.AddResource(dev_obj, "create", &resource.RouteResource {
						Name: obj2_ip + "-" + obj1_ip,
						Destination: obj2_ip,
						Device: "#" + dev.Status.Ip, // Todo: how to get net1
						Table: "cnf", // Todo: need check
					})

					log.Println("NAT Rule in " + strs[0] + " to " + obj2.Metadata.Name )
					resutil.AddResource(dev_obj, "create", &resource.FirewallNatResource {
						Name: obj2_ip + "-" + dev.Metadata.Name,
						DestinationIP: obj2_ip,
						Dest: "#source",
						SourceDestIP: dev.Status.DataIps[hubName],
						Index: "1",
						Target: "SNAT",
					})

					log.Println("Route Rule in " + obj2_ip + " to " + dev.Metadata.Name)
					resutil.AddResource(obj2, "create", &resource.RouteResource {
						Name: dev.Status.DataIps[hubName] + "-" + obj1_ip,
						Destination: dev.Status.DataIps[hubName],
						Device: "#" + obj2.Status.Ip,
						Table: "cnf",
					})

					log.Println("NAT Rule in " + obj2_ip + " to " + dev.Metadata.Name )
					resutil.AddResource(obj2, "create", &resource.FirewallNatResource {
						Name: dev.Status.DataIps[hubName] + "-" + obj2.Metadata.Name,
						DestinationIP: dev.Status.DataIps[hubName],
						Dest: "#source",
						SourceDestIP: obj2_ip,
						Index: "1",
						Target: "SNAT",
					})

				} else {
					log.Println("error in getting device")
					log.Println(err)
				}
			}
		}

		if is_delegated {
			log.Println("adding rules for delegate connections")
			resutil.AddResource(m2, "create", &resource.RouteResource {
				Name: "default4" + obj2.Metadata.Name,
				Destination: "default",
				Gateway: obj1_ip,
				Device: "#" + obj2_ip,
				Table: "cnf",
			})
			resutil.AddResource(m2, "create", &resource.FirewallNatResource {
				Name: "default4" + obj2.Metadata.Name,
				SourceDestIP: obj2_ip,
				Dest: "#source",
				Index: "0",
				Target: "SNAT",
			})
		}
	case DEVICETODEVICE:
		obj1 := m1.(*module.DeviceObject)
		obj2 := m2.(*module.DeviceObject)

		obj1_ip = obj1.Status.Ip
		obj2_ip = obj2.Status.Ip

		//Keypair
		obj1_crt, obj1_key, err := GetDeviceCertificate(m[OverlayResource], obj1.Metadata.Name)
		if err != nil {
			return err
		}
		obj2_crt, obj2_key, err := GetDeviceCertificate(m[OverlayResource], obj2.Metadata.Name)
		if err != nil {
			return err
		}

		conn := resource.Connection{
			Name:           DEFAULT_CONN + format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			ConnectionType: CONN_TYPE,
			Mode:           START_MODE,
			Mark:           DEFAULT_MARK,
			LocalUpDown:    DEFAULT_UPDOWN,
			CryptoProposal: all_proposals,
		}
		obj1_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj1.Metadata.Name, obj2.Metadata.Name),
			Type:                 POLICY_MODE,
			Remote:               obj2_ip,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(obj1_crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(obj1_key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj1.GetCertName(),
			RemoteIdentifier:     "CN=" + obj2.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          conn,
		}
		obj2_ipsec_resource = resource.IpsecResource{
			Name:                 format_resource_name(obj2.Metadata.Name, obj1.Metadata.Name),
			Type:                 POLICY_MODE,
			Remote:               obj1_ip,
			AuthenticationMethod: PUBKEY_AUTH,
			PublicCert:           base64.StdEncoding.EncodeToString([]byte(obj2_crt)),
			PrivateCert:          base64.StdEncoding.EncodeToString([]byte(obj2_key)),
			SharedCA:             base64.StdEncoding.EncodeToString([]byte(root_ca)),
			LocalIdentifier:      "CN=" + obj2.GetCertName(),
			RemoteIdentifier:     "CN=" + obj1.GetCertName(),
			CryptoProposal:       all_proposals,
			ForceCryptoProposal:  FORCECRYPTOPROPOSAL,
			Connections:          conn,
		}
	default:
		return pkgerrors.New("Unknown connection type")
	}

	resutil.AddResource(m1, "create", &obj1_ipsec_resource)
	resutil.AddResource(m2, "create", &obj2_ipsec_resource)

	log.Println("cend1 ip:", obj1_ip)
	log.Println("cend2 ip:", obj2_ip)
	cend1 := module.NewConnectionEnd(m1, obj1_ip)
	cend2 := module.NewConnectionEnd(m2, obj2_ip)
	co := module.NewConnectionObject(cend1, cend2)

	cm := GetConnectionManager()
	err = cm.Deploy(m[OverlayResource], co, resutil)
	if err != nil {
		return pkgerrors.Wrap(err, "Unable to create the object: fail to deploy resource")
	}

	return nil
}

func (c *OverlayObjectManager) DeleteConnection(m map[string]string, conn module.ConnectionObject) error {
	// use connection object to get connection ends
	// check if one of the ends is device object
	// if end1 yes, free ip with end2's name
	co1, _ := module.GetObjectBuilder().ToObject(conn.Info.End1.ConnObject)
	co2, _ := module.GetObjectBuilder().ToObject(conn.Info.End2.ConnObject)

	//Error: the re-constructed obj doesn't obtain the status
	if co1.GetType() == "Device" {
		log.Println("Enter Delete Connection with device on co1...")
		device_mgr := GetManagerset().Device
		device_mgr.FreeIP(m, co1, module.CreateEndName(co2.GetType(), co2.GetMetadata().Name))
	}

	if co2.GetType() == "Device" {
		log.Println("Enter Delete Connection with device on co2...")
		device_mgr := GetManagerset().Device
		device_mgr.FreeIP(m, co2, module.CreateEndName(co1.GetType(), co1.GetMetadata().Name))
	}

	conn_manager := GetConnectionManager()
	err := conn_manager.Undeploy(m[OverlayResource], conn)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *OverlayObjectManager) DeleteConnections(m map[string]string, m1 module.ControllerObject) error {
	//Get all connections related to the ControllerObject and do deletion^M
	conn_manager := GetConnectionManager()
	overlay_name := m[OverlayResource]
	conns, err := conn_manager.GetObjects(overlay_name, module.CreateEndName(m1.GetType(), m1.GetMetadata().Name))
	if err != nil {
		log.Println(err)
		return err
	} else {
		for i := 0; i < len(conns); i++ {
			conn := conns[i].(*module.ConnectionObject)
			err = c.DeleteConnection(m, *conn)
			if err != nil {
				log.Println("Failed to delete connection" + conn.GetMetadata().Name)
			}
		}
	}
	return nil
}
