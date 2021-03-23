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

import (
	"strconv"
    pkgerrors "github.com/pkg/errors"
)

const (
    MinProxyPort = 10000
    MaxProxyPort = 16000
)

// App contains metadata for Apps
type HubObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	Specification HubObjectSpec `json:"spec"`
	Status HubObjectStatus `json:"-"`
}

//HubObjectSpec contains the parameters
type HubObjectSpec struct {
	PublicIps    	[]string 	`json:"publicIps"`
	CertificateId 	string 		`json:"certificateId"`
	KubeConfig 		string 		`json:"kubeConfig"`
}

//HubObjectStatus
type HubObjectStatus struct {
	    Ip              string
        Data            map[string]string
        // Allocated proxy port for device
        ProxyPort       map[string]string
}

func (c *HubObject) GetMetadata() ObjectMetaData {
	return c.Metadata
}


func (c *HubObject) GetCertName() string {
    return "hub-" + c.Metadata.Name + "-cert"
}

func (c *HubObject) GetType() string {
	return "Hub"
}

func (c *HubObject) IsProxyPortUsed(port int) bool {
	_, ok := c.Status.ProxyPort[strconv.Itoa(port)]
	return ok
}

func (c *HubObject) SetProxyPort(port int, device string) {
	c.Status.ProxyPort[strconv.Itoa(port)] = device
}

func (c *HubObject) UnsetProxyPort(port int) {
	delete(c.Status.ProxyPort, strconv.Itoa(port))
}

func (c *HubObject) GetProxyPort(port int) string {
	return c.Status.ProxyPort[strconv.Itoa(port)]
}

func (c *HubObject) AllocateProxyPort() (int, error) {
	for i:=MinProxyPort; i<MaxProxyPort; i++ {
		if !c.IsProxyPortUsed(i) {
			return i, nil
		}
	}

	return 0, pkgerrors.New("Fail to allocate proxy port")
}

func init() {
	GetObjectBuilder().Register("Hub", &HubObject{})
}
