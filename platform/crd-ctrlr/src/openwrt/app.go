// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	appBaseURL = "sdewan/application/v1/"
)

type AppClient struct {
	OpenwrtClient *openwrtClient
}

// App Info
type SdewanApp struct {
	Name   string `json:"name"`
	IpList string `json:"iplist"`
}

type SdewanApps struct {
	Apps []SdewanApp `json:"apps"`
}

func (o *SdewanApp) GetName() string {
	return o.Name
}

// App APIs
// get apps
func (m *AppClient) GetApps() (*SdewanApps, error) {
	response, err := m.OpenwrtClient.Get(appBaseURL + "applications")
	if err != nil {
		return nil, err
	}

	var sdewanApps SdewanApps
	err2 := json.Unmarshal([]byte(response), &sdewanApps)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanApps, nil
}

// get app
func (m *AppClient) GetApp(app_name string) (*SdewanApp, error) {
	response, err := m.OpenwrtClient.Get(appBaseURL + "applications/" + app_name)
	if err != nil {
		return nil, err
	}

	var sdewanApp SdewanApp
	err2 := json.Unmarshal([]byte(response), &sdewanApp)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanApp, nil
}

// create app
func (m *AppClient) CreateApp(app SdewanApp) (*SdewanApp, error) {
	app_obj, _ := json.Marshal(app)
	response, err := m.OpenwrtClient.Post(appBaseURL+"applications/", string(app_obj))
	if err != nil {
		return nil, err
	}

	var sdewanApp SdewanApp
	err2 := json.Unmarshal([]byte(response), &sdewanApp)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanApp, nil
}

// delete app
func (m *AppClient) DeleteApp(app_name string) error {
	_, err := m.OpenwrtClient.Delete(appBaseURL + "applications/" + app_name)
	if err != nil {
		return err
	}

	return nil
}

// update app
func (m *AppClient) UpdateApp(app SdewanApp) (*SdewanApp, error) {
	app_obj, _ := json.Marshal(app)
	app_name := app.Name
	response, err := m.OpenwrtClient.Put(appBaseURL+"applications/"+app_name, string(app_obj))
	if err != nil {
		return nil, err
	}

	var sdewanApp SdewanApp
	err2 := json.Unmarshal([]byte(response), &sdewanApp)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanApp, nil
}
