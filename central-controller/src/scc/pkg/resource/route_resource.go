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

type RouteResource struct {
	Name        string
	Destination string
	Gateway     string
	Device      string
	Table       string
}

func (c *RouteResource) GetName() string {
	return c.Name
}

func (c *RouteResource) GetType() string {
	return "Route"
}

func (c *RouteResource) ToYaml(target string) string {
	basic := `apiVersion: ` + SdewanApiVersion + `
kind: CNFRoute
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
    targetCluster: ` + target + `
spec:
  dst: ` + c.Destination + `
  dev: "` + c.Device + `"
  table: ` + c.Table

	if c.Gateway != "" {
		basic += `
  gw: ` + c.Gateway
	}
	return basic
}

func init() {
	GetResourceBuilder().Register("Route", &RouteResource{})
}
