package openwrt

import (
    "openwrt"
    "reflect"
    "testing"
    "flag"
    "encoding/json"
    "fmt"
    "os"
)

var mwan3 openwrt.Mwan3Client
var service openwrt.ServiceClient
var available_policy string
var available_interface string
var available_interfaceb string

func TestMain(m *testing.M) {
    servIp := flag.String("ip", "10.244.0.18", "SDEWAN CNF Management IP Address")
    flag.Parse()

    client := openwrt.NewOpenwrtClient(*servIp, "root", "")
    mwan3 = openwrt.Mwan3Client{client}
    service = openwrt.ServiceClient{client}
    available_policy = "balanced"
    available_interface = "wan"
    available_interfaceb = "wanb"

    os.Exit(m.Run())
}

// Error handler
func handleError(t *testing.T, err error, name string, expectedErr bool, errorCode int) {
    if (err != nil) {
        if (expectedErr) {
            switch err.(type) {
            case *openwrt.OpenwrtError:
                if(errorCode != err.(*openwrt.OpenwrtError).Code) {
                    t.Errorf("Test case '%s': expected '%d', but got '%d'", name, errorCode, err.(*openwrt.OpenwrtError).Code)
                } else {
                    fmt.Printf("%s\n", err.(*openwrt.OpenwrtError).Message)
                }
            default:
                t.Errorf("Test case '%s': expected openwrt.OpenwrtError, but got '%s'", name, reflect.TypeOf(err).String())
            }
        } else {
            t.Errorf("Test case '%s': expected success, but got '%s'", name, reflect.TypeOf(err).String())
        }
    } else {
        if (expectedErr) {
            t.Errorf("Test case '%s': expected error code '%d', but success", name, errorCode)
        }
    }
}

func printError(err error) {
    switch err.(type) {
    case *openwrt.OpenwrtError:
        fmt.Printf("%s\n", err.(*openwrt.OpenwrtError).Message)
    default:
        fmt.Printf("%s\n", reflect.TypeOf(err).String())
    }
}

// Service API Test
func TestGetServices(t *testing.T) {
    res, _ := service.GetAvailableServices()

    if res == nil {
        t.Errorf("Test case GetServices: no available services")
        return
    }

    if !reflect.DeepEqual(res.Services, []string{"mwan3", "firewall", "ipsec"}) {
        t.Errorf("Test case GetServices: error available services returned")
    }
}

func TestExecuteService(t *testing.T) {
    tcases := []struct {
        name string
        service string
        action string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "Foo_Service",
            service: "foo_service",
            action: "restart",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Foo_action",
            service: "mwan3",
            action: "foo_action",
            expectedErr: true,
            expectedErrCode: 400,
        },
    }
    for _, tcase := range tcases {
        _, err := service.ExecuteService(tcase.service, tcase.action)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

// MWAN3 Rule API Test
func TestGetInterfaceStatus(t *testing.T) {
    res, _ := mwan3.GetInterfaceStatus()
    if res == nil {
        t.Errorf("Test case GetInterfaceStatus: can not get interfaces status")
    }
}

func TestGetRules(t *testing.T) {
    res, _ := mwan3.GetRules()
    if res == nil {
        t.Errorf("Test case GetRules: can not get mwan3 rules")
        return
    }

    if len(res.Rules) == 0 {
        t.Errorf("Test case GetRules: no rule defined")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetRule(t *testing.T) {
    tcases := []struct {
        name string
        rule string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetAvailableRule",
            rule: "default_rule",
        },
        {
            name: "GetFoolRule",
            rule: "foo_rule",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.GetRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolRule(t *testing.T) {
    tcases := []struct {
        name string
        rule openwrt.SdewanRule
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolPolicy",
            rule: openwrt.SdewanRule{Name:" ", Policy: "fool_policy"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_src_ip",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, SrcIp: "10.200.200.500"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_src_port",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, SrcPort: "-1"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_dest_ip",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, DestIp: "10.200"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_dest_port",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, DestPort: "0"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_proto",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, Proto: "fool_protocol"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_family",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, Family: "fool_family"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_sticky",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, Sticky: "2"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Wrong_timeout",
            rule: openwrt.SdewanRule{Name:" ", Policy: available_policy, Timeout: "-1"},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.CreateRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolRule(t *testing.T) {
    tcases := []struct {
        name string
        rule string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            rule: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            rule: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        err := mwan3.DeleteRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolRule(t *testing.T) {
    tcases := []struct {
        name string
        rule openwrt.SdewanRule
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            rule: openwrt.SdewanRule{Name:"fool_name", Policy: available_policy},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.UpdateRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestRule(t *testing.T) {
    var err error
    var ret_rule *openwrt.SdewanRule

    var rule_name = "test_rule"
    var rule = openwrt.SdewanRule{Name:rule_name, Policy: available_policy, SrcIp: "10.10.10.10/24", SrcPort: "80"}
    var update_rule = openwrt.SdewanRule{Name:rule_name, Policy: available_policy, SrcIp: "100.100.100.100/24", SrcPort: "8000", DestIp: "172.172.172.172/16", DestPort: "8080"}

    _, err = mwan3.GetRule(rule_name)
    if (err == nil) {
        err = mwan3.DeleteRule(rule_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestRule: failed to delete rule '%s'", rule_name)
            return
        }
    }

    // Create rule
    ret_rule, err = mwan3.CreateRule(rule)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to create rule '%s'", rule_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_rule)
        fmt.Printf("Created Rule: %s\n", string(p_data))
    }

    ret_rule, err = mwan3.GetRule(rule_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to get created rule")
        return
    } else {
        if( ret_rule.SrcPort != "80" ) {
            t.Errorf("Test case TestRule: failed to create rule")
            return
        }
    }

    // Update rule
    ret_rule, err = mwan3.UpdateRule(update_rule)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to update rule '%s'", rule_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_rule)
        fmt.Printf("Updated Rule: %s\n", string(p_data))
    }

    ret_rule, err = mwan3.GetRule(rule_name)
    if (err != nil) {
        t.Errorf("Test case TestRule: failed to get updated rule")
        return
    } else {
        if( ret_rule.SrcIp != "100.100.100.100/24" ) {
            t.Errorf("Test case TestRule: failed to update rule")
            return
        }
    }

    // Delete rule
    err = mwan3.DeleteRule(rule_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to delete rule '%s'", rule_name)
        return
    }

    ret_rule, err = mwan3.GetRule(rule_name)
    if (err == nil) {
        t.Errorf("Test case TestRule: failed to delete rule")
        return
    }
}

// MWAN3 Policy API Tests
func TestGetPolicies(t *testing.T) {
    res, _ := mwan3.GetPolicies()
    if res == nil {
        t.Errorf("Test case GetPolicies: can not get mwan3 policies")
        return
    }

    if len(res.Policies) == 0 {
        t.Errorf("Test case GetPolicies: no policy defined")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetPolicy(t *testing.T) {
    tcases := []struct {
        name string
        policy string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetAvailablePolicy",
            policy: available_policy,
        },
        {
            name: "GetFoolPolicy",
            policy: "foo_policy",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.GetPolicy(tcase.policy)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolPolicy(t *testing.T) {
    tcases := []struct {
        name string
        policy openwrt.SdewanPolicy
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            policy: openwrt.SdewanPolicy{Name:" ", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"1", Weight:"2"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "EmptyMember",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Policy-Conflict",
            policy: openwrt.SdewanPolicy{Name:"wan_only", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"1", Weight:"2"}}},
            expectedErr: true,
            expectedErrCode: 409,
        },
        {
            name: "WrongMember-interface",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:""}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-interface-name",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:"fool-name", Metric:"1", Weight:"2"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-duplicate-interface",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"1", Weight:"2"}, {Interface:available_interface, Metric:"2", Weight:"3"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-Metric",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:""}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-Weight",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"2"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-Metric",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"-1", Weight:"2"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "WrongMember-Weight",
            policy: openwrt.SdewanPolicy{Name:"policy1", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"2", Weight:"0"}}},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.CreatePolicy(tcase.policy)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolPolicy(t *testing.T) {
    tcases := []struct {
        name string
        policy string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            policy: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            policy: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
        {
            name: "Delete-InUsePolicy",
            policy: available_policy,
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        err := mwan3.DeletePolicy(tcase.policy)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolPolicy(t *testing.T) {
    tcases := []struct {
        name string
        policy openwrt.SdewanPolicy
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            policy: openwrt.SdewanPolicy{Name:"fool_name", Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"1", Weight:"2"}}},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := mwan3.UpdatePolicy(tcase.policy)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestPolicy(t *testing.T) {
    var err error
    var ret_policy *openwrt.SdewanPolicy

    var policy_name = "test_policy"
    var policy = openwrt.SdewanPolicy{Name:policy_name, Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"1", Weight:"2"}}}
    var update_policy = openwrt.SdewanPolicy{Name:policy_name, Members:[]openwrt.SdewanMember{{Interface:available_interface, Metric:"2", Weight:"3"}, {Interface:available_interfaceb, Metric:"2", Weight:"3"}}}

    _, err = mwan3.GetPolicy(policy_name)
    if (err == nil) {
        err = mwan3.DeletePolicy(policy_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestPolicy: failed to delete policy '%s'", policy_name)
            return
        }
    }

    // Create policy
    ret_policy, err = mwan3.CreatePolicy(policy)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestPolicy: failed to create policy '%s'", policy_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_policy)
        fmt.Printf("Created Policy: %s\n", string(p_data))
    }

    ret_policy, err = mwan3.GetPolicy(policy_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestPolicy: failed to get created policy")
        return
    } else {
        if( len(ret_policy.Members) != 1 || ret_policy.Members[0].Metric != "1" ) {
            t.Errorf("Test case TestPolicy: failed to create policy")
            return
        }
    }

    // Update policy
    ret_policy, err = mwan3.UpdatePolicy(update_policy)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestPolicy: failed to update policy '%s'", policy_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_policy)
        fmt.Printf("Updated Policy: %s\n", string(p_data))
    }

    ret_policy, err = mwan3.GetPolicy(policy_name)
    if (err != nil) {
        t.Errorf("Test case TestPolicy: failed to get updated policy")
        return
    } else {
        if( len(ret_policy.Members) != 2 || ret_policy.Members[1].Metric != "2" ) {
            t.Errorf("Test case TestPolicy: failed to update policy")
            return
        }
    }

    // Delete policy
    err = mwan3.DeletePolicy(policy_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestPolicy: failed to delete policy '%s'", policy_name)
        return
    }

    ret_policy, err = mwan3.GetPolicy(policy_name)
    if (err == nil) {
        t.Errorf("Test case TestPolicy: failed to delete policy")
        return
    }
}

