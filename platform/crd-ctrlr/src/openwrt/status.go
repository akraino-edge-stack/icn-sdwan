// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	statusBaseURL = "sdewan/v1/"
)

type StatusClient struct {
	OpenwrtClient *openwrtClient
}

// Status Info
type SdewanModuleStatus struct {
	Name   string      `json:"name"`
	Status interface{} `json:"status"`
}

func (o *SdewanModuleStatus) GetName() string {
	return o.Name
}

// Status APIs
// get status
func (m *StatusClient) GetStatus() (*[]SdewanModuleStatus, error) {
	response, err := m.OpenwrtClient.Get(statusBaseURL + "status")
	if err != nil {
		return nil, err
	}

	var sdewanStatus []SdewanModuleStatus
	err = json.Unmarshal([]byte(response), &sdewanStatus)
	if err != nil {
		return nil, err
	}

	return &sdewanStatus, nil
}

func (m *StatusClient) GetModuleStatus(mname string) (*SdewanModuleStatus, error) {
	response, err := m.OpenwrtClient.Get(statusBaseURL + "status/" + mname)
	if err != nil {
		return nil, err
	}

	var sdewanModuleStatus SdewanModuleStatus
	err = json.Unmarshal([]byte(response), &sdewanModuleStatus)
	if err != nil {
		return nil, err
	}

	return &sdewanModuleStatus, nil
}
