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

import (
)

type FirewallDnatResource struct {
    Name string
    Source string
    SourceIP string
    SourceDestIP string
    SourceDestPort string
    DestinationIP string
    DestinationPort string
    Protocol string
}

func (c *FirewallDnatResource) GetName() string {
    return c.Name
}

func (c *FirewallDnatResource) GetType() string {
    return "FirewallDNAT"
}

func (c *FirewallDnatResource) ToYaml() string {
    basic := `apiVersion: ` + SdewanApiVersion + `
kind: FirewallDNAT
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
spec:
  src: ` + c.Source + `
  src_dip: ` + c.SourceDestIP + `
  src_dport: ` + c.SourceDestPort + `
  dest_ip: ` + c.DestinationIP + `
  dest_port: ` + c.DestinationPort + `
  proto: ` + c.Protocol + `
  target: DNAT `

    if c.SourceIP != "" {
      basic +=  `
  src_ip: ` + c.SourceIP
  }

  return basic
}

func init() {
  GetResourceBuilder().Register("FirewallDnat", &FirewallDnatResource{})
}
