// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	svcBaseURL = "sdewan/service/v1/"
)

type SvcClient struct {
	OpenwrtClient *openwrtClient
}

// Svc Info
type SdewanSvc struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Port     string `json:"port"`
	DPort    string `json:"dport"`
}

type SdewanSvcs struct {
	Svcs []SdewanSvc `json:"svcs"`
}

func (o *SdewanSvc) GetName() string {
	return o.Name
}

// Svc APIs
// get svcs
func (m *SvcClient) GetSvcs() (*SdewanSvcs, error) {
	response, err := m.OpenwrtClient.Get(svcBaseURL + "services")
	if err != nil {
		return nil, err
	}

	var sdewanSvcs SdewanSvcs
	err2 := json.Unmarshal([]byte(response), &sdewanSvcs)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanSvcs, nil
}

// get svc
func (m *SvcClient) GetSvc(svc_name string) (*SdewanSvc, error) {
	response, err := m.OpenwrtClient.Get(svcBaseURL + "services/" + svc_name)
	if err != nil {
		return nil, err
	}

	var sdewanSvc SdewanSvc
	err2 := json.Unmarshal([]byte(response), &sdewanSvc)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanSvc, nil
}

// create svc
func (m *SvcClient) CreateSvc(svc SdewanSvc) (*SdewanSvc, error) {
	svc_obj, _ := json.Marshal(svc)
	response, err := m.OpenwrtClient.Post(svcBaseURL+"services/", string(svc_obj))
	if err != nil {
		return nil, err
	}

	var sdewanSvc SdewanSvc
	err2 := json.Unmarshal([]byte(response), &sdewanSvc)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanSvc, nil
}

// delete svc
func (m *SvcClient) DeleteSvc(svc_name string) error {
	_, err := m.OpenwrtClient.Delete(svcBaseURL + "services/" + svc_name)
	if err != nil {
		return err
	}

	return nil
}

// update svc
func (m *SvcClient) UpdateSvc(svc SdewanSvc) (*SdewanSvc, error) {
	svc_obj, _ := json.Marshal(svc)
	svc_name := svc.Name
	response, err := m.OpenwrtClient.Put(svcBaseURL+"services/"+svc_name, string(svc_obj))
	if err != nil {
		return nil, err
	}

	var sdewanSvc SdewanSvc
	err2 := json.Unmarshal([]byte(response), &sdewanSvc)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanSvc, nil
}
