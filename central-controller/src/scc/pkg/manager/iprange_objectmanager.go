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
	"encoding/json"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/infra/validation"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/go-playground/validator/v10"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"io"
)

type IPRangeObjectKey struct {
	OverlayName string `json:"overlay-name"`
	IPRangeName string `json:"iprange-name"`
}

type ProviderIPRangeObjectKey struct {
	IPRangeName string `json:"iprange-name"`
}

// IPRangeObjectManager implements the ControllerObjectManager
type IPRangeObjectManager struct {
	BaseObjectManager
	provider bool
}

func NewIPRangeObjectManager(provider bool) *IPRangeObjectManager {
	object_meta := "iprange"
	if provider {
		object_meta = "provider-" + object_meta
	}

	validate := validation.GetValidator(object_meta)
	validate.RegisterStructValidation(ValidateIPRangeObject, module.IPRangeObject{})

	return &IPRangeObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        object_meta,
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
		provider,
	}
}

func ValidateIPRangeObject(sl validator.StructLevel) {
	obj := sl.Current().Interface().(module.IPRangeObject)

	if obj.Specification.MinIp != 0 && obj.Specification.MaxIp != 0 {
		if obj.Specification.MinIp > obj.Specification.MaxIp {
			sl.ReportError(obj.Specification.MinIp, "Range", "Range", "InValidateIPRange", "")
		}
	}
}

func (c *IPRangeObjectManager) GetResourceName() string {
	return IPRangeResource
}

func (c *IPRangeObjectManager) IsOperationSupported(oper string) bool {
	if oper == "PUT" {
		// Not allowed for update
		return false
	}
	return true
}

func (c *IPRangeObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.IPRangeObject{}
}

func (c *IPRangeObjectManager) SetIPRangeName(k db.Key, name string) {
	if c.provider {
		ko := k.(*ProviderIPRangeObjectKey)
		ko.IPRangeName = name
	} else {
		ko := k.(*IPRangeObjectKey)
		ko.IPRangeName = name
	}
}

func (c *IPRangeObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	var key db.Key

	if c.provider {
		key = &ProviderIPRangeObjectKey{
			IPRangeName: "",
		}
	} else {
		overlay_name := m[OverlayResource]
		key = &IPRangeObjectKey{
			OverlayName: overlay_name,
			IPRangeName: "",
		}
	}

	if isCollection == true {
		return key, nil
	}

	to := t.(*module.IPRangeObject)
	meta_name := to.Metadata.Name
	res_name := m[IPRangeResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		c.SetIPRangeName(key, res_name)
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		c.SetIPRangeName(key, meta_name)
	}

	return key, nil
}

func (c *IPRangeObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.IPRangeObject
	err := json.NewDecoder(r).Decode(&v)

	// initial Status
	for i := 0; i < 32; i++ {
		v.Status.Masks[i] = 0
	}
	v.Status.Data = make(map[string]string)
	return &v, err
}

func (c *IPRangeObjectManager) GetDefinedObjects(m map[string]string) ([]module.ControllerObject, error) {
	objs, err := c.GetObjects(m)
	if err != nil {
		return []module.ControllerObject{}, pkgerrors.Wrap(err, "Failed to get available IPRange objects")
	}

	if c.provider {
		ipr_manager := GetManagerset().IPRange
		overlay_manager := GetManagerset().Overlay

		// concact ip ranges defined in all overlays
		ol_objs, err := overlay_manager.GetObjects(m)
		if err != nil {
			return []module.ControllerObject{}, pkgerrors.Wrap(err, "Failed to get overlays")
		}

		for _, ol_obj := range ol_objs {
			o_m := make(map[string]string)
			o_m[OverlayResource] = ol_obj.GetMetadata().Name
			// get ip range for the overlay
			ip_objs, err := ipr_manager.GetObjects(o_m)
			if err != nil {
				return []module.ControllerObject{}, pkgerrors.Wrap(err, "Failed to get ip ranges for overlay")
			}
			objs = append(objs, ip_objs...)
		}
	} else {
		// concact ip ranges defined in provider
		providerIP_manager := GetManagerset().ProviderIPRange
		p_objs, err := providerIP_manager.GetObjects(m)
		if err != nil {
			return []module.ControllerObject{}, pkgerrors.Wrap(err, "Failed to get provider IPRange objects")
		}

		objs = append(objs, p_objs...)
	}

	return objs, nil
}

func (c *IPRangeObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// Check whether conflict with other IPRange object
	objs, err := c.GetDefinedObjects(m)
	if err != nil {
		return t, pkgerrors.Wrap(err, "Failed to get available IPRange objects")
	}

	ot := t.(*module.IPRangeObject)
	for _, obj := range objs {
		if ot.IsConflict(obj.(*module.IPRangeObject)) {
			return c.CreateEmptyObject(), pkgerrors.New("Conflicted with IPRange object: " + obj.(*module.IPRangeObject).Metadata.Name)
		}
	}

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)

	return t, err
}

func (c *IPRangeObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	return t, err
}

func (c *IPRangeObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *IPRangeObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().UpdateObject(c, m, t)

	return t, err
}

func (c *IPRangeObjectManager) DeleteObject(m map[string]string) error {
	// Check whether in used
	obj, err := c.GetObject(m)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get IPRange object")
	}

	if obj.(*module.IPRangeObject).InUsed() {
		return pkgerrors.New("The IPRange object is in used")
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)

	return err
}

func (c *IPRangeObjectManager) Allocate(oname string, name string) (string, error) {
	m := make(map[string]string)

	if !c.provider {
		m[OverlayResource] = oname
	}

	objs, err := c.GetObjects(m)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed to get available IPRange objects")
	}

	for _, obj := range objs {
		tobj := obj.(*module.IPRangeObject)
		aip, err := tobj.Allocate(name)
		if err == nil {
			// save update object in DB
			c.UpdateObject(m, tobj)
			return aip, nil
		}
	}

	return "", pkgerrors.New("No available ip")
}

func (c *IPRangeObjectManager) Free(oname string, ip string) error {
	m := make(map[string]string)

	if !c.provider {
		m[OverlayResource] = oname
	}

	objs, err := c.GetObjects(m)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get available IPRange objects")
	}

	for _, obj := range objs {
		tobj := obj.(*module.IPRangeObject)
		err := tobj.Free(ip)
		if err == nil {
			// save update object in DB
			c.UpdateObject(m, tobj)
			return nil
		}
	}

	return pkgerrors.New("ip " + ip + " is not allocated")
}

func (c *IPRangeObjectManager) FreeAll(oname string) error {
	m := make(map[string]string)

	if !c.provider {
		m[OverlayResource] = oname
	}

	objs, err := c.GetObjects(m)
	if err != nil {
		return pkgerrors.Wrap(err, "Failed to get available IPRange objects")
	}

	for _, obj := range objs {
		tobj := obj.(*module.IPRangeObject)
		err := tobj.FreeAll()
		if err == nil {
			// save update object in DB
			c.UpdateObject(m, tobj)
		}
	}

	return nil
}
