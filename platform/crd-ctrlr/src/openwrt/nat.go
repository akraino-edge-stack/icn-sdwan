// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package openwrt

import (
	"encoding/json"
)

const (
	natBaseURL = "sdewan/nat/v1/"
)

type NatClient struct {
	OpenwrtClient *openwrtClient
}

// Nat
type SdewanNat struct {
	Name     string `json:"name"`
	Src      string `json:"src"`
	SrcIp    string `json:"src_ip"`
	SrcDIp   string `json:"src_dip"`
	SrcPort  string `json:"src_port"`
	SrcDPort string `json:"src_dport"`
	Proto    string `json:"proto"`
	Dest     string `json:"dest"`
	DestIp   string `json:"dest_ip"`
	DestPort string `json:"dest_port"`
	Target   string `json:"target"`
	Index    string `json:"index"`
}

func (o *SdewanNat) GetName() string {
	return o.Name
}

func (o *SdewanNat) SetFullName(namespace string) {
	o.Name = namespace + o.Name
}

type SdewanNats struct {
	Nats []SdewanNat `json:"nats"`
}

// Nat APIs
// get nats
func (f *NatClient) GetNats() (*SdewanNats, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(natBaseURL + "nats")
	if err != nil {
		return nil, err
	}

	var sdewanNats SdewanNats
	err = json.Unmarshal([]byte(response), &sdewanNats)
	if err != nil {
		return nil, err
	}

	return &sdewanNats, nil
}

// get nat
func (m *NatClient) GetNat(nat string) (*SdewanNat, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(natBaseURL + "nats/" + nat)
	if err != nil {
		return nil, err
	}

	var sdewanNat SdewanNat
	err = json.Unmarshal([]byte(response), &sdewanNat)
	if err != nil {
		return nil, err
	}

	return &sdewanNat, nil
}

// create nat
func (m *NatClient) CreateNat(nat SdewanNat) (*SdewanNat, error) {
	var response string
	var err error
	nat_obj, _ := json.Marshal(nat)
	response, err = m.OpenwrtClient.Post(natBaseURL+"nats", string(nat_obj))
	if err != nil {
		return nil, err
	}

	var sdewanNat SdewanNat
	err = json.Unmarshal([]byte(response), &sdewanNat)
	if err != nil {
		return nil, err
	}

	return &sdewanNat, nil
}

// delete nat
func (m *NatClient) DeleteNat(nat_name string) error {
	_, err := m.OpenwrtClient.Delete(natBaseURL + "nats/" + nat_name)
	if err != nil {
		return err
	}

	return nil
}

// update nat
func (m *NatClient) UpdateNat(nat SdewanNat) (*SdewanNat, error) {
	var response string
	var err error
	nat_obj, _ := json.Marshal(nat)
	nat_name := nat.Name
	response, err = m.OpenwrtClient.Put(natBaseURL+"nats/"+nat_name, string(nat_obj))
	if err != nil {
		return nil, err
	}

	var sdewanNat SdewanNat
	err = json.Unmarshal([]byte(response), &sdewanNat)
	if err != nil {
		return nil, err
	}

	return &sdewanNat, nil
}
