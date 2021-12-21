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
	"encoding/json"
	pkgerrors "github.com/pkg/errors"
	"reflect"
	"strings"
)

type ResourceBuilder struct {
	omap map[string]reflect.Type
}

var res_builder = ResourceBuilder{
	omap: make(map[string]reflect.Type),
}

func GetResourceBuilder() *ResourceBuilder {
	return &res_builder
}

func (c *ResourceBuilder) Register(name string, r interface{}) {
	c.omap[name] = reflect.TypeOf(r).Elem()
}

func (c *ResourceBuilder) ToString(obj ISdewanResource) (string, error) {
	obj_str, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return obj.GetType() + "-" + string(obj_str), nil
}

func (c *ResourceBuilder) ToObject(obj_str string) (ISdewanResource, error) {
	if !strings.Contains(obj_str, "-") {
		return &EmptyResource{"", ""}, pkgerrors.New("Not a valid object")
	}
	strs := strings.SplitN(obj_str, "-", 2)
	if len(strs) != 2 {
		return &EmptyResource{"", ""}, pkgerrors.New("Not a valid object")
	}

	if v, ok := c.omap[strs[0]]; ok {
		retObj := reflect.New(v).Interface()
		err := json.Unmarshal([]byte(strs[1]), retObj)
		return retObj.(ISdewanResource), err
	} else {
		return &EmptyResource{"", ""}, pkgerrors.New("Not a valid object")
	}
}
