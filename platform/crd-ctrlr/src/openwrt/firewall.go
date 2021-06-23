// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	firewallBaseURL = "sdewan/firewall/v1/"
)

type FirewallClient struct {
	OpenwrtClient *openwrtClient
}

// Firewall Zones
type SdewanFirewallZone struct {
	Name             string   `json:"name"`
	Network          []string `json:"network"`
	Masq             string   `json:"masq"`
	MasqSrc          []string `json:"masq_src"`
	MasqDest         []string `json:"masq_dest"`
	MasqAllowInvalid string   `json:"masq_allow_invalid"`
	MtuFix           string   `json:"mtu_fix"`
	Input            string   `json:"input"`
	Forward          string   `json:"forward"`
	Output           string   `json:"output"`
	Family           string   `json:"family"`
	Subnet           []string `json:"subnet"`
	ExtraSrc         string   `json:"extra_src"`
	ExtraDest        string   `json:"etra_dest"`
}

func (o *SdewanFirewallZone) GetName() string {
	return o.Name
}

type SdewanFirewallZones struct {
	Zones []SdewanFirewallZone `json:"zones"`
}

// Firewall Forwarding
type SdewanFirewallForwarding struct {
	Name   string `json:"name"`
	Src    string `json:"src"`
	Dest   string `json:"dest"`
	Family string `json:"family"`
}

func (o *SdewanFirewallForwarding) GetName() string {
	return o.Name
}

type SdewanFirewallForwardings struct {
	Forwardings []SdewanFirewallForwarding `json:"forwardings"`
}

// Firewall Rule
type SdewanFirewallRule struct {
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

func (o *SdewanFirewallRule) GetName() string {
	return o.Name
}

type SdewanFirewallRules struct {
	Rules []SdewanFirewallRule `json:"rules"`
}

// Firewall Redirect
type SdewanFirewallRedirect struct {
	Name     string `json:"name"`
	Src      string `json:"src"`
	SrcIp    string `json:"src_ip"`
	SrcDIp   string `json:"src_dip"`
	SrcMac   string `json:"src_mac"`
	SrcPort  string `json:"src_port"`
	SrcDPort string `json:"src_dport"`
	Proto    string `json:"proto"`
	Dest     string `json:"dest"`
	DestIp   string `json:"dest_ip"`
	DestPort string `json:"dest_port"`
	Mark     string `json:"mark"`
	Target   string `json:"target"`
	Family   string `json:"family"`
}

func (o *SdewanFirewallRedirect) GetName() string {
	return o.Name
}

type SdewanFirewallRedirects struct {
	Redirects []SdewanFirewallRedirect `json:"redirects"`
}

// Zone APIs
// get zones
func (f *FirewallClient) GetZones() (*SdewanFirewallZones, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(firewallBaseURL + "zones")
	if err != nil {
		return nil, err
	}

	var sdewanFirewallZones SdewanFirewallZones
	err = json.Unmarshal([]byte(response), &sdewanFirewallZones)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallZones, nil
}

// get zone
func (m *FirewallClient) GetZone(zone string) (*SdewanFirewallZone, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(firewallBaseURL + "zones/" + zone)
	if err != nil {
		return nil, err
	}

	var sdewanFirewallZone SdewanFirewallZone
	err = json.Unmarshal([]byte(response), &sdewanFirewallZone)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallZone, nil
}

// create zone
func (m *FirewallClient) CreateZone(zone SdewanFirewallZone) (*SdewanFirewallZone, error) {
	var response string
	var err error
	zone_obj, _ := json.Marshal(zone)
	response, err = m.OpenwrtClient.Post(firewallBaseURL+"zones", string(zone_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallZone SdewanFirewallZone
	err = json.Unmarshal([]byte(response), &sdewanFirewallZone)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallZone, nil
}

// delete zone
func (m *FirewallClient) DeleteZone(zone_name string) error {
	_, err := m.OpenwrtClient.Delete(firewallBaseURL + "zones/" + zone_name)
	if err != nil {
		return err
	}

	return nil
}

// update zone
func (m *FirewallClient) UpdateZone(zone SdewanFirewallZone) (*SdewanFirewallZone, error) {
	var response string
	var err error
	zone_obj, _ := json.Marshal(zone)
	zone_name := zone.Name
	response, err = m.OpenwrtClient.Put(firewallBaseURL+"zones/"+zone_name, string(zone_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallZone SdewanFirewallZone
	err = json.Unmarshal([]byte(response), &sdewanFirewallZone)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallZone, nil
}

// Rule APIs
// get rules
func (f *FirewallClient) GetRules() (*SdewanFirewallRules, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(firewallBaseURL + "rules")
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRules SdewanFirewallRules
	err = json.Unmarshal([]byte(response), &sdewanFirewallRules)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRules, nil
}

// get rule
func (m *FirewallClient) GetRule(rule string) (*SdewanFirewallRule, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(firewallBaseURL + "rules/" + rule)
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRule SdewanFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRule, nil
}

// create rule
func (m *FirewallClient) CreateRule(rule SdewanFirewallRule) (*SdewanFirewallRule, error) {
	var response string
	var err error
	rule_obj, _ := json.Marshal(rule)
	response, err = m.OpenwrtClient.Post(firewallBaseURL+"rules", string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRule SdewanFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRule, nil
}

// delete rule
func (m *FirewallClient) DeleteRule(rule_name string) error {
	_, err := m.OpenwrtClient.Delete(firewallBaseURL + "rules/" + rule_name)
	if err != nil {
		return err
	}

	return nil
}

// update rule
func (m *FirewallClient) UpdateRule(rule SdewanFirewallRule) (*SdewanFirewallRule, error) {
	var response string
	var err error
	rule_obj, _ := json.Marshal(rule)
	rule_name := rule.Name
	response, err = m.OpenwrtClient.Put(firewallBaseURL+"rules/"+rule_name, string(rule_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRule SdewanFirewallRule
	err = json.Unmarshal([]byte(response), &sdewanFirewallRule)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRule, nil
}

// Forwarding APIs
// get forwardings
func (f *FirewallClient) GetForwardings() (*SdewanFirewallForwardings, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(firewallBaseURL + "forwardings")
	if err != nil {
		return nil, err
	}

	var sdewanFirewallForwardings SdewanFirewallForwardings
	err = json.Unmarshal([]byte(response), &sdewanFirewallForwardings)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallForwardings, nil
}

// get forwarding
func (m *FirewallClient) GetForwarding(forwarding string) (*SdewanFirewallForwarding, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(firewallBaseURL + "forwardings/" + forwarding)
	if err != nil {
		return nil, err
	}

	var sdewanFirewallForwarding SdewanFirewallForwarding
	err = json.Unmarshal([]byte(response), &sdewanFirewallForwarding)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallForwarding, nil
}

// create forwarding
func (m *FirewallClient) CreateForwarding(forwarding SdewanFirewallForwarding) (*SdewanFirewallForwarding, error) {
	var response string
	var err error
	forwarding_obj, _ := json.Marshal(forwarding)
	response, err = m.OpenwrtClient.Post(firewallBaseURL+"forwardings", string(forwarding_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallForwarding SdewanFirewallForwarding
	err = json.Unmarshal([]byte(response), &sdewanFirewallForwarding)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallForwarding, nil
}

// delete forwarding
func (m *FirewallClient) DeleteForwarding(forwarding_name string) error {
	_, err := m.OpenwrtClient.Delete(firewallBaseURL + "forwardings/" + forwarding_name)
	if err != nil {
		return err
	}

	return nil
}

// update forwarding
func (m *FirewallClient) UpdateForwarding(forwarding SdewanFirewallForwarding) (*SdewanFirewallForwarding, error) {
	var response string
	var err error
	forwarding_obj, _ := json.Marshal(forwarding)
	forwarding_name := forwarding.Name
	response, err = m.OpenwrtClient.Put(firewallBaseURL+"forwardings/"+forwarding_name, string(forwarding_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallForwarding SdewanFirewallForwarding
	err = json.Unmarshal([]byte(response), &sdewanFirewallForwarding)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallForwarding, nil
}

// Redirect APIs
// get redirects
func (f *FirewallClient) GetRedirects() (*SdewanFirewallRedirects, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(firewallBaseURL + "redirects")
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRedirects SdewanFirewallRedirects
	err = json.Unmarshal([]byte(response), &sdewanFirewallRedirects)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRedirects, nil
}

// get redirect
func (m *FirewallClient) GetRedirect(redirect string) (*SdewanFirewallRedirect, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(firewallBaseURL + "redirects/" + redirect)
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRedirect SdewanFirewallRedirect
	err = json.Unmarshal([]byte(response), &sdewanFirewallRedirect)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRedirect, nil
}

// create redirect
func (m *FirewallClient) CreateRedirect(redirect SdewanFirewallRedirect) (*SdewanFirewallRedirect, error) {
	var response string
	var err error
	redirect_obj, _ := json.Marshal(redirect)
	response, err = m.OpenwrtClient.Post(firewallBaseURL+"redirects", string(redirect_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRedirect SdewanFirewallRedirect
	err = json.Unmarshal([]byte(response), &sdewanFirewallRedirect)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRedirect, nil
}

// delete redirect
func (m *FirewallClient) DeleteRedirect(redirect_name string) error {
	_, err := m.OpenwrtClient.Delete(firewallBaseURL + "redirects/" + redirect_name)
	if err != nil {
		return err
	}

	return nil
}

// update redirect
func (m *FirewallClient) UpdateRedirect(redirect SdewanFirewallRedirect) (*SdewanFirewallRedirect, error) {
	var response string
	var err error
	redirect_obj, _ := json.Marshal(redirect)
	redirect_name := redirect.Name
	response, err = m.OpenwrtClient.Put(firewallBaseURL+"redirects/"+redirect_name, string(redirect_obj))
	if err != nil {
		return nil, err
	}

	var sdewanFirewallRedirect SdewanFirewallRedirect
	err = json.Unmarshal([]byte(response), &sdewanFirewallRedirect)
	if err != nil {
		return nil, err
	}

	return &sdewanFirewallRedirect, nil
}
