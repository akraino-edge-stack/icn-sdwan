// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	networkfirewallBaseURL = "sdewan/networkfirewall/v1/"
)

type NetworkFirewallClient struct {
	OpenwrtClient *openwrtClient
}

// NetworkFirewall Rule
type SdewanNetworkFirewallRule struct {
	Name     string   `json:"name"`
	Src      string   `json:"src"`
	SrcIp    string   `json:"src_ip"`
	SrcMac   string   `json:"src_mac"`
	SrcPort  string   `json:"src_port"`
	Proto    string   `json:"proto"`
	IcmpType []string `json:"icmp_type"`
	Dest     string   `json:"dest"`
	DestIp   string   `json:"dest_ip"`
	DestPort string   `json:"dest_port"`
	Mark     string   `json:"mark"`
	Target   string   `json:"target"`
	SetMark  string   `json:"set_mark"`
	SetXmark string   `json:"set_xmark"`
	Family   string   `json:"family"`
	Extra    string   `json:"extra"`
}

func (o *SdewanNetworkFirewallRule) GetName() string {
	return o.Name
}

type SdewanNetworkFirewallRules struct {
	Rules []SdewanNetworkFirewallRule `json:"rules"`
}

// get rules
func (f *NetworkFirewallClient) GetRules() (*SdewanNetworkFirewallRules, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(networkfirewallBaseURL + "rules")
	if err != nil {
		return nil, err
	}

	var sdewanNetworkFirewallRules SdewanNetworkFirewallRules
	err = json.Unmarshal([]byte(response), &sdewanNetworkFirewallRules)
	if err != nil {
		return nil, err
	}

	return &sdewanNetworkFirewallRules, nil
}

// get rule
func (m *NetworkFirewallClient) GetRule(rule string) (*SdewanNetworkFirewallRule, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(networkfirewallBaseURL + "rules/" + rule)
	if err != nil {
		return nil, err
	}

	var sdewanNetworkFirewallRule SdewanNetworkFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanNetworkFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanNetworkFirewallRule, nil
}

// create rule
func (m *NetworkFirewallClient) CreateRule(rule SdewanNetworkFirewallRule) (*SdewanNetworkFirewallRule, error) {
	var response string
	var err error
	rule_obj, _ := json.Marshal(rule)
	response, err = m.OpenwrtClient.Post(networkfirewallBaseURL+"rules", string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanNetworkFirewallRule SdewanNetworkFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanNetworkFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanNetworkFirewallRule, nil
}

// delete rule
func (m *NetworkFirewallClient) DeleteRule(rule_name string) error {
	_, err := m.OpenwrtClient.Delete(networkfirewallBaseURL + "rules/" + rule_name)
	if err != nil {
		return err
	}

	return nil
}

// update rule
func (m *NetworkFirewallClient) UpdateRule(rule SdewanNetworkFirewallRule) (*SdewanNetworkFirewallRule, error) {
	var response string
	var err error
	rule_obj, _ := json.Marshal(rule)
	rule_name := rule.Name
	response, err = m.OpenwrtClient.Put(networkfirewallBaseURL+"rules/"+rule_name, string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanNetworkFirewallRule SdewanNetworkFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanNetworkFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanNetworkFirewallRule, nil
}