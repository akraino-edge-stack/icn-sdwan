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
	"encoding/json"
)

const (
	serviceBaseURL = "sdewan/v1/"
)

var available_Services = []string{"mwan3", "firewall", "ipsec"}

type ServiceClient struct {
	OpenwrtClient *openwrtClient
}

// Service API struct
type AvailableServices struct {
	Services []string `json:"services"`
}

// get available services
func (s *ServiceClient) GetAvailableServices() (*AvailableServices, error) {
	response, err := s.OpenwrtClient.Get(serviceBaseURL + "services")
	if err != nil {
		return nil, err
	}

	var servs AvailableServices
	err2 := json.Unmarshal([]byte(response), &servs)
	if err2 != nil {
		return nil, err2
	}

	return &servs, nil
}

func (s *ServiceClient) formatExecuteServiceBody(operation string) string {
	return "{\"action\":\"" + operation + "\"}"
}

// execute operation on service
func (s *ServiceClient) ExecuteService(service string, operation string) (bool, error) {
	if !IsContained(available_Services, service) {
		return false, &OpenwrtError{Code: 400, Message: "Bad Request: not supported service(" + service + ")"}
	}

	_, err := s.OpenwrtClient.Put(serviceBaseURL+"service/"+service, s.formatExecuteServiceBody(operation))
	if err != nil {
		return false, err
	}

	return true, nil
}
