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

package openwrt

import (
	"reflect"
)

// util function to check whether items contains item
func IsContained(items interface{}, item interface{}) bool {
	switch reflect.TypeOf(items).Kind() {
	case reflect.Slice:
		v := reflect.ValueOf(items)
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(item, v.Index(i).Interface()) {
				return true
			}
		}
	default:
		return false
	}

	return false
}
