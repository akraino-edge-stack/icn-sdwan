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

type ProposalResource struct {
	Name       string
	Encryption string
	Hash       string
	DhGroup    string
}

func (c *ProposalResource) GetName() string {
	return c.Name
}

func (c *ProposalResource) GetType() string {
	return "Proposal"
}

func (c *ProposalResource) ToYaml(target string) string {
	return `apiVersion: ` + SdewanApiVersion + `
kind: IpsecProposal
metadata:
  name: ` + c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
    targetCluster: ` + target + `
spec:
  encryption_algorithm: ` + c.Encryption + `
  hash_algorithm: ` + c.Hash + `
  dh_group: ` + c.DhGroup
}

func init() {
	GetResourceBuilder().Register("Proposal", &ProposalResource{})
}
