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

const (
	MinProxyPort = 10000
	MaxProxyPort = 16000
)

// App contains metadata for Apps
type HubObject struct {
	Metadata      ObjectMetaData  `json:"metadata"`
	Specification HubObjectSpec   `json:"spec"`
	Status        HubObjectStatus `json:"-"`
}

//HubObjectSpec contains the parameters
type HubObjectSpec struct {
	PublicIps     []string `json:"publicIps"`
	CertificateId string   `json:"certificateId"`
	KubeConfig    string   `json:"kubeConfig" encrypted:""`
}

//HubObjectStatus
type HubObjectStatus struct {
	Ip   string
	Data map[string]string
	// Devices that this hub delegates
	DelegateDevices []string
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

func init() {
	GetObjectBuilder().Register("Hub", &HubObject{})
}
