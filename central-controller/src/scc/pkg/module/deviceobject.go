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

package module

// App contains metadata for Apps
type DeviceObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	Specification DeviceObjectSpec `json:"spec"`
	Status DeviceObjectStatus `json:"-"`
}

// DeviceObjectSpec contains the parameters
type DeviceObjectSpec struct {
	PublicIps    	[]string 	`json:"publicIps"`
	ForceHubConnectivity    	bool 	`json:"forceHubConnectivity"`
	ProxyHub		string 	`json:"proxyHub"`
	ProxyHubPort	int 	`json:"proxyHubPort"`
	UseHub4Internet	bool 	`json:"useHub4Internet"`
	DedicatedSFC	bool 	`json:"dedicatedSFC"`	
	CertificateId 	string 	`json:"certificateId"`
	KubeConfig 		string 	`json:"kubeConfig"`
}

// DeviceObjectStatus
type DeviceObjectStatus struct {
    // 1: use public ip 2: use hub as proxy
    Mode  int
    // ip used for external connection
    // if Mode=1, ip is one of public ip
    // if Mode=2, ip is the OIP allocated by SCC
    Ip string
    // Status Data
    Data map[string]string
}

func (c *DeviceObject) GetMetadata() ObjectMetaData {
	return c.Metadata
}

func (c *DeviceObject) GetType() string {
	return "Device"
}

func (c *DeviceObject) IsProxyHub(hub_name string) bool {
	if c.Status.Mode == 2 {
        return c.Specification.ProxyHub == hub_name
	}

    return false
}

func init() {
	GetObjectBuilder().Register("Device", &DeviceObject{})
}

func (c *DeviceObject) GetCertName() string {
    return "device-" + c.Metadata.Name + "-cert"
}
