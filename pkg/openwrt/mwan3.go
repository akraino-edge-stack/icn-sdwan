package openwrt

import (
    "encoding/json"
)

const (
    mwan3BaseURL = "sdewan/mwan3/v1/"
)

type Mwan3Client struct {
    OpenwrtClient *openwrtClient
}

// MWAN3 interface status
type IpStatus struct {
    Status string `json:"status"`
    Latency int `json:"latency"`
    Packetloss int `json:"packetloss"`
    Ip string `json:"ip"`
}

type WanInterfaceStatus struct {
    Running bool `json:"running"`
    Score int `json:"score"`
    Lost int `json:"lost"`
    Status string `json:"status"`
    Age int `json:"age"`
    Turn int `json:"turn"`
    Ips []IpStatus `json:"track_ip"`
}

type InterfaceStatus struct {
    Interfaces map[string]WanInterfaceStatus `json:"interfaces"`
    Connected map[string][]string `json:"connected"`
}

// MWAN3 Policy
type SdewanMember struct {
    Interface string `json:"interface"`
    Metric string `json:"metric"`
    Weight string `json:"weight"`
}

type SdewanPolicy struct {
    Name string `json:"name"`
    Members []SdewanMember `json:"members"`
}

type SdewanPolicies struct {
    Policies []SdewanPolicy `json:"policies"`
}

// MWAN3 Rule
type SdewanRule struct {
    Name string `json:"name"`
    Policy string `json:"policy"`
    SrcIp string `json:"src_ip"`
    SrcPort string `json:"src_port"`
    DestIp string `json:"dest_ip"`
    DestPort string `json:"dest_port"`
    Proto string `json:"proto"`
    Family string `json:"family"`
    Sticky string `json:"sticky"`
    Timeout string `json:"timeout"`
}

type SdewanRules struct {
    Rules []SdewanRule `json:"rules"`
}

// get interface status
func (m *Mwan3Client) GetInterfaceStatus() (*InterfaceStatus, error) {
    response, err := m.OpenwrtClient.Get("admin/status/mwan/interface_status")
    if (err != nil) {
        return nil, err
    }

    var interfaceStatus InterfaceStatus
    err2 := json.Unmarshal([]byte(response), &interfaceStatus)
    if (err2 != nil) {
        return nil, err2
    }

    return &interfaceStatus, nil
}

// Policy APIs
// get policies
func (m *Mwan3Client) GetPolicies() (*SdewanPolicies, error) {
    response, err := m.OpenwrtClient.Get(mwan3BaseURL + "policies")
    if (err != nil) {
        return nil, err
    }

    var sdewanPolicies SdewanPolicies
    err2 := json.Unmarshal([]byte(response), &sdewanPolicies)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanPolicies, nil
}

// get policy
func (m *Mwan3Client) GetPolicy(policy_name string) (*SdewanPolicy, error) {
    response, err := m.OpenwrtClient.Get(mwan3BaseURL + "policy/" + policy_name)
    if (err != nil) {
        return nil, err
    }

    var sdewanPolicy SdewanPolicy
    err2 := json.Unmarshal([]byte(response), &sdewanPolicy)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanPolicy, nil
}

// create policy
func (m *Mwan3Client) CreatePolicy(policy SdewanPolicy) (*SdewanPolicy, error) {
    policy_obj, _ := json.Marshal(policy)
    response, err := m.OpenwrtClient.Post(mwan3BaseURL + "policy", string(policy_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanPolicy SdewanPolicy
    err2 := json.Unmarshal([]byte(response), &sdewanPolicy)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanPolicy, nil
}

// delete policy
func (m *Mwan3Client) DeletePolicy(policy_name string) (error) {
    _, err := m.OpenwrtClient.Delete(mwan3BaseURL + "policy/" + policy_name)
    if (err != nil) {
        return err
    }

    return nil
}

// update policy
func (m *Mwan3Client) UpdatePolicy(policy SdewanPolicy) (*SdewanPolicy, error) {
    policy_obj, _ := json.Marshal(policy)
    policy_name := policy.Name
    response, err := m.OpenwrtClient.Put(mwan3BaseURL + "policy/" + policy_name, string(policy_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanPolicy SdewanPolicy
    err2 := json.Unmarshal([]byte(response), &sdewanPolicy)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanPolicy, nil
}

// Rule APIs
// get rules
func (m *Mwan3Client) GetRules() (*SdewanRules, error) {
    response, err := m.OpenwrtClient.Get(mwan3BaseURL + "rules")
    if (err != nil) {
        return nil, err
    }

    var sdewanRules SdewanRules
    err2 := json.Unmarshal([]byte(response), &sdewanRules)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanRules, nil
}

// get rule
func (m *Mwan3Client) GetRule(rule string) (*SdewanRule, error) {
    response, err := m.OpenwrtClient.Get(mwan3BaseURL + "rule/" + rule)
    if (err != nil) {
        return nil, err
    }

    var sdewanRule SdewanRule
    err2 := json.Unmarshal([]byte(response), &sdewanRule)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanRule, nil
}

// create rule
func (m *Mwan3Client) CreateRule(rule SdewanRule) (*SdewanRule, error) {
    rule_obj, _ := json.Marshal(rule)
    response, err := m.OpenwrtClient.Post(mwan3BaseURL + "rule", string(rule_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanRule SdewanRule
    err2 := json.Unmarshal([]byte(response), &sdewanRule)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanRule, nil
}

// delete rule
func (m *Mwan3Client) DeleteRule(rule_name string) (error) {
    _, err := m.OpenwrtClient.Delete(mwan3BaseURL + "rule/" + rule_name)
    if (err != nil) {
        return err
    }

    return nil
}

// update rule
func (m *Mwan3Client) UpdateRule(rule SdewanRule) (*SdewanRule, error) {
    rule_obj, _ := json.Marshal(rule)
    rule_name := rule.Name
    response, err := m.OpenwrtClient.Put(mwan3BaseURL + "rule/" + rule_name, string(rule_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanRule SdewanRule
    err2 := json.Unmarshal([]byte(response), &sdewanRule)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanRule, nil
}