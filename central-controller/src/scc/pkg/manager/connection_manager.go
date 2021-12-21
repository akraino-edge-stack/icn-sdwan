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

package manager

import (
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
	"log"
)

type ConnectionManager struct {
	storeName string
	tagMeta   string
}

type ConnectionKey struct {
	OverlayName string `json:"overlay-name"`
	End1        string `json:"end1-name"`
	End2        string `json:"end2-name"`
}

var connutil = ConnectionManager{
	storeName: StoreName,
	tagMeta:   "connection",
}

func GetConnectionManager() *ConnectionManager {
	return &connutil
}

func (c *ConnectionManager) CreateEmptyObject() module.ControllerObject {
	return &module.ConnectionObject{}
}

func (c *ConnectionManager) GetStoreName() string {
	return c.storeName
}

func (c *ConnectionManager) GetStoreMeta() string {
	return c.tagMeta
}

func (c *ConnectionManager) Deploy(overlay string, cm module.ConnectionObject, resutil *ResUtil) error {
/*	resutil := NewResUtil()

	// add resource for End1
	co1, _ := module.GetObjectBuilder().ToObject(cm.Info.End1.ConnObject)
	for _, r_str := range cm.Info.End1.Resources {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co1, "create", r)
	}
	for _, r_str := range cm.Info.End1.ReservedRes {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co1, "create", r)
	}

	// add resource for End2
	co2, _ := module.GetObjectBuilder().ToObject(cm.Info.End2.ConnObject)
	for _, r_str := range cm.Info.End2.Resources {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co2, "create", r)
	}
	for _, r_str := range cm.Info.End2.ReservedRes {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co2, "create", r)
	}
*/
	// add resource to cm
	rm := resutil.GetResources()
	for device, res := range rm {
		for _, resource := range res.Resources {
			cm.Info.AddResource(device, resource.Resource.GetName(), resource.Resource.GetType())
		}
	}

	// Deploy resources
	err := resutil.Deploy(overlay, cm.Metadata.Name, "YAML")

	if err != nil {
		log.Println(err)
		cm.Info.State = module.StateEnum.Error
		cm.Info.ErrorMessage = err.Error()
	} else {
		cm.Info.State = module.StateEnum.Deployed
	}

	log.Println(cm.Info.End1.IP)
	// Save to DB
	_, err = c.UpdateObject(overlay, cm)

	return err
}

func (c *ConnectionManager) Undeploy(overlay string, cm module.ConnectionObject) error {
	resutil := NewResUtil()
/*
	// add resource for End1 (reservedRes will be kept)
	co1, _ := module.GetObjectBuilder().ToObject(cm.Info.End1.ConnObject)
	for _, r_str := range cm.Info.End1.Resources {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co1, "create", r)
	}

	// add resource for End2 (reservedRes will be kept)
	co2, _ := module.GetObjectBuilder().ToObject(cm.Info.End2.ConnObject)
	for _, r_str := range cm.Info.End2.Resources {
		r, _ := resource.GetResourceBuilder().ToObject(r_str)
		resutil.AddResource(co2, "create", r)
	}
*/
	// fill resutil
	for _, res := range cm.Info.Resources {
		co, _ := module.GetObjectBuilder().ToObject(res.ConnObject)
		resutil.AddResource(co, "delete", &resource.EmptyResource{res.Name, res.Type})
	}

	// Undeploy resources
	err := resutil.Undeploy(overlay)

	if err != nil {
		log.Println(err)
		cm.Info.State = module.StateEnum.Error
		cm.Info.ErrorMessage = err.Error()
	} else {
		cm.Info.State = module.StateEnum.Undeployed
	}

	// Delete connection object
	err = c.DeleteObject(overlay, cm.Info.End1.Name, cm.Info.End2.Name)

	return err
}

func (c *ConnectionManager) UpdateObject(overlay string, cm module.ConnectionObject) (module.ControllerObject, error) {
	key := ConnectionKey{
		OverlayName: overlay,
		End1:        cm.Info.End1.Name,
		End2:        cm.Info.End2.Name,
	}

	err := db.DBconn.Insert(c.GetStoreName(), key, nil, c.GetStoreMeta(), cm)
	if err != nil {
		return c.CreateEmptyObject(), pkgerrors.Wrap(err, "Unable to create the object")
	}
	return &cm, err
}

func (c *ConnectionManager) GetObject(overlay string, key1 string, key2 string) (module.ControllerObject, error) {
	key := ConnectionKey{
		OverlayName: overlay,
		End1:        key1,
		End2:        key2,
	}
	value, err := db.DBconn.Find(c.GetStoreName(), key, c.GetStoreMeta())
	if err != nil {
		return c.CreateEmptyObject(), err
	}

	if value == nil {
		key = ConnectionKey{
			OverlayName: overlay,
			End1:        key2,
			End2:        key1,
		}
		value, err = db.DBconn.Find(c.GetStoreName(), key, c.GetStoreMeta())
		if err != nil {
			return c.CreateEmptyObject(), err
		}
	}

	if value != nil {
		r := c.CreateEmptyObject()
		err = db.DBconn.Unmarshal(value[0], r)
		if err != nil {
			return c.CreateEmptyObject(), pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return r, nil
	}

	return c.CreateEmptyObject(), pkgerrors.New("No Object")
}

func (c *ConnectionManager) GetObjects(overlay string, key string) ([]module.ControllerObject, error) {
	key1 := ConnectionKey{
		OverlayName: overlay,
		End1:        key,
		End2:        "",
	}
	key2 := ConnectionKey{
		OverlayName: overlay,
		End1:        "",
		End2:        key,
	}

	var resp []module.ControllerObject

	// find objects with end1=key
	values, err := db.DBconn.Find(c.GetStoreName(), key1, c.GetStoreMeta())
	if err != nil {
		return []module.ControllerObject{}, pkgerrors.Wrap(err, "Get Overlay Objects")
	}

	for _, value := range values {
		t := c.CreateEmptyObject()
		err = db.DBconn.Unmarshal(value, t)
		if err != nil {
			return []module.ControllerObject{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, t)
	}

	// find objects with end2=key
	values, err = db.DBconn.Find(c.GetStoreName(), key2, c.GetStoreMeta())
	if err != nil {
		return []module.ControllerObject{}, pkgerrors.Wrap(err, "Get Overlay Objects")
	}

	for _, value := range values {
		t := c.CreateEmptyObject()
		err = db.DBconn.Unmarshal(value, t)
		if err != nil {
			return []module.ControllerObject{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, t)
	}

	return resp, nil
}

func (c *ConnectionManager) DeleteObject(overlay string, key1 string, key2 string) error {
	key := ConnectionKey{
		OverlayName: overlay,
		End1:        key1,
		End2:        key2,
	}

	err := db.DBconn.Remove(c.GetStoreName(), key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Object")
	}

	return err
}
