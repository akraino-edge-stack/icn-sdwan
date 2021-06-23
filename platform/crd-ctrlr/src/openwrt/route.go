// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	routeBaseURL = "sdewan/route/v1/"
)

type RouteClient struct {
	OpenwrtClient *openwrtClient
}

// Route Info
type SdewanRoute struct {
	Name  string `json:"name"`
	Dst   string `json:"dst"`
	Gw    string `json:"gw"`
	Dev   string `json:"dev"`
	Table string `json:"table"`
}

type SdewanRoutes struct {
	Routes []SdewanRoute `json:"routes"`
}

func (o *SdewanRoute) GetName() string {
	return o.Name
}

// Route APIs
// get routes
func (m *RouteClient) GetRoutes() (*SdewanRoutes, error) {
	response, err := m.OpenwrtClient.Get(routeBaseURL + "routes")
	if err != nil {
		return nil, err
	}

	var sdewanRoutes SdewanRoutes
	err2 := json.Unmarshal([]byte(response), &sdewanRoutes)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRoutes, nil
}

// get route
func (m *RouteClient) GetRoute(route_name string) (*SdewanRoute, error) {
	response, err := m.OpenwrtClient.Get(routeBaseURL + "routes/" + route_name)
	if err != nil {
		return nil, err
	}

	var sdewanRoute SdewanRoute
	err2 := json.Unmarshal([]byte(response), &sdewanRoute)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRoute, nil
}

// create route
func (m *RouteClient) CreateRoute(route SdewanRoute) (*SdewanRoute, error) {
	route_obj, _ := json.Marshal(route)
	response, err := m.OpenwrtClient.Post(routeBaseURL+"routes/", string(route_obj))
	if err != nil {
		return nil, err
	}

	var sdewanRoute SdewanRoute
	err2 := json.Unmarshal([]byte(response), &sdewanRoute)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRoute, nil
}

// delete route
func (m *RouteClient) DeleteRoute(route_name string) error {
	_, err := m.OpenwrtClient.Delete(routeBaseURL + "routes/" + route_name)
	if err != nil {
		return err
	}

	return nil
}

// update route
func (m *RouteClient) UpdateRoute(route SdewanRoute) (*SdewanRoute, error) {
	route_obj, _ := json.Marshal(route)
	route_name := route.Name
	response, err := m.OpenwrtClient.Put(routeBaseURL+"routes/"+route_name, string(route_obj))
	if err != nil {
		return nil, err
	}

	var sdewanRoute SdewanRoute
	err2 := json.Unmarshal([]byte(response), &sdewanRoute)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRoute, nil
}
