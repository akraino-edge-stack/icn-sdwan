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

type FirewallNatResource struct {
	Name            string
	Source          string
	SourceIP        string
	SourcePort      string
	SourceDestIP    string
	SourceDestPort  string
	Dest            string
	DestinationIP   string
	DestinationPort string
	Protocol        string
	Target          string
	Index           string
}

func (c *FirewallNatResource) GetName() string {
	return c.Name
}

func (c *FirewallNatResource) GetType() string {
	return "FirewallNAT"
}

func (c *FirewallNatResource) ToYaml(target string) string {
	basic := `apiVersion: ` + SdewanApiVersion + `
kind: CNFNAT
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
    targetCluster: ` + target + `
spec:
  target: ` + c.Target + `
  src_dip: ` + c.SourceDestIP
        if c.DestinationIP != "" {
		basic += `
  dest_ip: ` + c.DestinationIP
        }

        if c.DestinationPort != "" {
		basic += `
  dest_port: ` + c.DestinationPort
        }
	if c.Dest != "" {
		basic += `
  dest: "` + c.Dest + `"`
        }
	if c.SourceDestPort != "" {
		basic += `
  src_dport: ` + c.SourceDestPort
        }
	if c.Protocol != "" {
		basic += `
  proto: ` + c.Protocol
	}
	if c.Source != "" {
		basic += `
  src: "` + c.Source + `"`
        }
	if c.SourceIP != "" {
		basic += `
  src_ip: ` + c.SourceIP
	}
	if c.Index != "" {
		basic += `
  index: "` + c.Index + `"`
        }

	return basic
}

func init() {
	GetResourceBuilder().Register("FirewallNat", &FirewallNatResource{})
}
