/*
 * Copyright 2021 Intel Corporation, Inc
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
	"strings"
)

func format_resource_name(name1 string, name2 string) string {
	name1 = strings.Replace(name1, "-", "", -1)
	name2 = strings.Replace(name2, "-", "", -1)

	return strings.ToLower(name1 + name2)
}

func format_ip_as_suffix(ip string) string {
	return "_" + strings.Replace(ip, ".", "", -1)
}
