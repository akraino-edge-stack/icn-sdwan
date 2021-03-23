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
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
)

// App contains metadata for Apps
type ProposalObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	Specification ProposalObjectSpec `json:"spec"`
}

//ProposalObjectSpec contains the parameters
type ProposalObjectSpec struct {
	Encryption    	string 	`json:"encryption"`
	Hash    		string 	`json:"hash"`
	DhGroup    		string 	`json:"dhGroup"`
}

func (c *ProposalObject) GetMetadata() ObjectMetaData {
	return c.Metadata
}

func (c *ProposalObject) GetType() string {
    return "Proposal"
}

func (c *ProposalObject) ToResource() *resource.ProposalResource {
    return &resource.ProposalResource{
        Name: c.Metadata.Name,
        Encryption: c.Specification.Encryption,
        Hash: c.Specification.Hash,
        DhGroup: c.Specification.DhGroup,
    }
}