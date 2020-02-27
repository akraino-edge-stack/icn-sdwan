package main

import (
    "openwrt"
    "fmt"
    "encoding/json"
    "strconv"
)

func testGetServices(serv *openwrt.ServiceClient) {
    res, err := serv.GetAvailableServices()
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        servs, err := json.Marshal(res)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(servs))
        }
    }
}

func testPutService(serv *openwrt.ServiceClient) {
    res, err := serv.ExecuteService("mwan3", "start")
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        fmt.Printf("Result: %s\n", strconv.FormatBool(res))
    }
}

func testGetInterfaceStatus(mwan3 *openwrt.Mwan3Client) {
    is, err := mwan3.GetInterfaceStatus()
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        fmt.Printf("Wan interface status: %s\n", is.Interfaces["wan"].Status)
        is_data, err := json.Marshal(is)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(is_data))
        }
    }

    is = nil
}

func testGetPolicies(mwan3 *openwrt.Mwan3Client) {
    res, err := mwan3.GetPolicies()
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        p_data, err := json.Marshal(res)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(p_data))
        }
    }
}

func testGetPolicy(mwan3 *openwrt.Mwan3Client, policy string) {
    res, err := mwan3.GetPolicy(policy)
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        p_data, err := json.Marshal(res)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(p_data))
        }
    }
}

func testGetRules(mwan3 *openwrt.Mwan3Client) {
    res, err := mwan3.GetRules()
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        p_data, err := json.Marshal(res)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(p_data))
        }
    }
}

func testGetRule(mwan3 *openwrt.Mwan3Client, rule string) {
    res, err := mwan3.GetRule(rule)
    if(err != nil) {
        switch err.(type) {
            case *openwrt.OpenwrtError:
                fmt.Printf("Openwrt Error: %s\n", err.Error())
                fmt.Printf("Openwrt Error: %d\n", err.(*openwrt.OpenwrtError).Code)
            default:
                fmt.Printf("%s\n", err.Error())
        }
    } else {
        p_data, err := json.Marshal(res)
        if(err != nil) {
            fmt.Printf("%s\n", err.Error())
        } else {
            fmt.Printf("%s\n", string(p_data))
        }
    }
}

func main() {
    client := openwrt.NewOpenwrtClient("10.244.0.18", "root", "")
    mwan3 := openwrt.Mwan3Client{OpenwrtClient: client}
    service := openwrt.ServiceClient{client}

    testGetServices(&service)

    testGetInterfaceStatus(&mwan3)
    testGetPolicies(&mwan3)
    testGetPolicy(&mwan3, "balanced")
    testGetPolicy(&mwan3, "foo_policy")

    testGetRules(&mwan3)
    testGetRule(&mwan3, "https")
    testGetRule(&mwan3, "foo_rule")
    // testPutService(&service)  
}
