// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	ruleBaseURL = "sdewan/rule/v1/"
)

type RouteRuleClient struct {
	OpenwrtClient *openwrtClient
}

// RouteRule Info
type SdewanRouteRule struct {
	Name   string `json:"name"`
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Flag   bool   `json:"flag"`
	Prio   string `json:"prio"`
	Fwmark string `json:"fwmark"`
	Table  string `json:"table"`
}

type SdewanRouteRules struct {
	RouteRules []SdewanRouteRule `json:"routerules"`
}

func (o *SdewanRouteRule) GetName() string {
	return o.Name
}
func (o *SdewanRouteRule) SetFullName(namespace string) {
	o.Name = namespace + o.Name
}

// RouteRule APIs
// get rules
func (m *RouteRuleClient) GetRouteRules() (*SdewanRouteRules, error) {
	response, err := m.OpenwrtClient.Get(ruleBaseURL + "rules")
	if err != nil {
		return nil, err
	}

	var sdewanRouteRules SdewanRouteRules
	err2 := json.Unmarshal([]byte(response), &sdewanRouteRules)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRouteRules, nil
}

// get rule
func (m *RouteRuleClient) GetRouteRule(rule_name string) (*SdewanRouteRule, error) {
	response, err := m.OpenwrtClient.Get(ruleBaseURL + "rules/" + rule_name)
	if err != nil {
		return nil, err
	}

	var sdewanRouteRule SdewanRouteRule
	err2 := json.Unmarshal([]byte(response), &sdewanRouteRule)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRouteRule, nil
}

// create rule
func (m *RouteRuleClient) CreateRouteRule(rule SdewanRouteRule) (*SdewanRouteRule, error) {
	rule_obj, _ := json.Marshal(rule)
	response, err := m.OpenwrtClient.Post(ruleBaseURL+"rules/", string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanRouteRule SdewanRouteRule
	err2 := json.Unmarshal([]byte(response), &sdewanRouteRule)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRouteRule, nil
}

// delete rule
func (m *RouteRuleClient) DeleteRouteRule(rule_name string) error {
	_, err := m.OpenwrtClient.Delete(ruleBaseURL + "rules/" + rule_name)
	if err != nil {
		return err
	}

	return nil
}

// update rule
func (m *RouteRuleClient) UpdateRouteRule(rule SdewanRouteRule) (*SdewanRouteRule, error) {
	rule_obj, _ := json.Marshal(rule)
	rule_name := rule.Name
	response, err := m.OpenwrtClient.Put(ruleBaseURL+"rules/"+rule_name, string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanRouteRule SdewanRouteRule
	err2 := json.Unmarshal([]byte(response), &sdewanRouteRule)
	if err2 != nil {
		return nil, err2
	}

	return &sdewanRouteRule, nil
}
