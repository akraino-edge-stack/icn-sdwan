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
	"log"
)

type states struct {
	Created    string
	Deployed   string
	Undeployed string
	Error      string
}

var StateEnum = &states{
	Created:    "Created",
	Deployed:   "Deployed",
	Undeployed: "Undeployed",
	Error:      "Error",
}

type ConnectionResource struct {
	ConnObject  string  `json:"-"`
	Name     	string 	`json:"-"`
	Type        string  `json:"-"`
}

type ConnectionObject struct {
	Metadata ObjectMetaData `json:"metadata"`
	Info     ConnectionInfo `json:"information"`
}

//ConnectionInfo contains the connection information
type ConnectionInfo struct {
	End1         ConnectionEnd `json:"end1"`
	End2         ConnectionEnd `json:"end2"`
//	ContextId    string        `json:"-"`
	Resources     []ConnectionResource `json:"-"`
	State        string        `json:"state"`
	ErrorMessage string        `json:"message"`
}

type ConnectionEnd struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	IP          string   `json:"ip"`
	ConnObject  string   `json:"-"`
//	Resources   []string `json:"-"`
//	ReservedRes []string `json:"-"`
}

func (c *ConnectionObject) GetMetadata() ObjectMetaData {
	return c.Metadata
}

func (c *ConnectionObject) GetType() string {
	return "Connection"
}

func (c *ConnectionObject) GetPeer(t string, n string) (string, string, string) {
	e1 := c.Info.End1
	e2 := c.Info.End2
	if e1.Type == t && e1.Name == CreateEndName(t, n) {
		return e2.Type, e2.Name, e2.IP
	} else {
		if e2.Type == t && e2.Name == CreateEndName(t, n) {
			return e1.Type, e1.Name, e1.IP
		}
	}

	return "", "", ""
}

func CreateEndName(t string, n string) string {
	return t + "." + n
}

func CreateConnectionName(e1 string, e2 string) string {
	return e1 + "-" + e2
}

func NewConnectionEnd(conn_obj ControllerObject, ip string) ConnectionEnd {
	obj_str, err := GetObjectBuilder().ToString(conn_obj)
	if err == nil {
		return ConnectionEnd{
			Name:        CreateEndName(conn_obj.GetType(), conn_obj.GetMetadata().Name),
			Type:        conn_obj.GetType(),
			IP:          ip,
			ConnObject:  obj_str,
//			Resources:   []string{},
//			ReservedRes: []string{},
		}
	} else {
		log.Println(err)
		return ConnectionEnd{}
	}
}

func NewConnectionObject(end1 ConnectionEnd, end2 ConnectionEnd) ConnectionObject {
	return ConnectionObject{
		Metadata: ObjectMetaData{CreateConnectionName(end1.Name, end2.Name), "", "", ""},
		Info: ConnectionInfo{
			End1:         end1,
			End2:         end2,
//			ContextId:    "",
			Resources:    []ConnectionResource{},
			State:        StateEnum.Created,
			ErrorMessage: "",
		},
	}
}

func (c *ConnectionInfo) AddResource(device ControllerObject, resource string, res_type string) {
	dev_str, err := GetObjectBuilder().ToString(device)
	if err == nil {
		c.Resources = append(c.Resources, ConnectionResource{dev_str, resource, res_type})
	} else {
		log.Println(err)
	}
}

/*
func (c *ConnectionEnd) contains(res resource.ISdewanResource, isReserved bool) bool {
	if isReserved {
		for _, r_str := range c.ReservedRes {
			r, err := resource.GetResourceBuilder().ToObject(r_str)
			if err == nil {
				if r.GetName() == res.GetName() &&
					r.GetType() == res.GetType() {
					return true
				}
			}
		}
	} else {
		for _, r_str := range c.Resources {
			r, err := resource.GetResourceBuilder().ToObject(r_str)
			if err == nil {
				if r.GetName() == res.GetName() &&
					r.GetType() == res.GetType() {
					return true
				}
			}
		}
	}

	return false
}

func (c *ConnectionEnd) AddResource(res resource.ISdewanResource, isReserved bool) error {
	if !c.contains(res, isReserved) {
		res_str, err := resource.GetResourceBuilder().ToString(res)
		if err == nil {
			if isReserved {
				c.ReservedRes = append(c.ReservedRes, res_str)
			} else {
				c.Resources = append(c.Resources, res_str)
			}
		}
	}

	return nil
}
*/