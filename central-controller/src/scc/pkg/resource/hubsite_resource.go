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

package resource

import ()

type HubSiteResource struct {
	Name      string
	Type      string
	Site      string
	Subnet    string
	HubIP     string
	DevicePIP string
}

func (c *HubSiteResource) GetName() string {
	return c.Name
}

func (c *HubSiteResource) GetType() string {
	return "HubSite"
}

func (c *HubSiteResource) ToYaml(target string) string {
	result := `apiVersion: ` + SdewanApiVersion + `
kind: CNFHubSite
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
    targetCluster: ` + target + `
spec:
  type: ` + c.Type
	if c.Site != "" {
		result += `
  site: '` + c.Site + `'`
	}
	if c.Subnet != "" {
		result += `
  subnet: '` + c.Subnet + `'`
	}
	if c.HubIP != "" {
		result += `
  hubip: '` + c.HubIP + `'`
	}
	if c.DevicePIP != "" {
		result += `
  devicepip: '` + c.DevicePIP + `'`
	}

	return result
}

func init() {
	GetResourceBuilder().Register("HubSite", &HubSiteResource{})
}
