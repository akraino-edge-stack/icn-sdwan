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
	"strings"
)

type FirewallZoneResource struct {
    Name string
    Network []string
    Input string
    Output string
    Forward string
    MASQ string
    MTU_FIX string
}

func (c *FirewallZoneResource) GetName() string {
    return c.Name
}

func (c *FirewallZoneResource) GetType() string {
    return "FirewallZone"
}

func (c *FirewallZoneResource) ToYaml() string {
    basic := `apiVersion: ` + SdewanApiVersion + `
kind: FirewallZone
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
spec:
  network: [` + strings.Join(c.Network, ",") + `]
  input: ` + c.Input + `
  output: ` + c.Output + `
  forward: ` + c.Forward

    if (c.MASQ != "" && c.MTU_FIX != "") {
      optional := `
  masq: ` + c.MASQ + `
  mtu_fix: ` + c.MTU_FIX
      basic += optional
    }

    return basic
}

func init() {
  GetResourceBuilder().Register("FirewallZone", &FirewallZoneResource{})
}
