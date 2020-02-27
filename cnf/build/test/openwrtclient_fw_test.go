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

var fw openwrt.FirewallClient
var available_zone_1 string
var available_zone_2 string

func TestMain(m *testing.M) {
    servIp := flag.String("ip", "10.244.0.18", "SDEWAN CNF Management IP Address")
    flag.Parse()

    client := openwrt.NewOpenwrtClient(*servIp, "root", "")
    fw = openwrt.FirewallClient{client}
    available_zone_1 = "wan"
    available_zone_2 = "lan"

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

// Firewall Zone API Test
func TestGetZones(t *testing.T) {
    res, err := fw.GetZones()
    if res == nil {
        printError(err)
        t.Errorf("Test case GetZones: can not get firewall zones")
        return
    }

    if len(res.Zones) == 0 {
        fmt.Printf("Test case GetZones: no zone defined")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetZone(t *testing.T) {
    tcases := []struct {
        name string
        zone string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetAvailableZone",
            zone: available_zone_1,
        },
        {
            name: "GetFoolRule",
            zone: "foo_zone",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.GetZone(tcase.zone)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolZone(t *testing.T) {
    tcases := []struct {
        name string
        zone openwrt.SdewanFirewallZone
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            zone: openwrt.SdewanFirewallZone{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidNetwork",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Network:[]string{"fool_name"}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidMasq",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Masq:"2"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidMasqSrc",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", MasqSrc:[]string{"256.0.0.0/0"}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidMasqDest",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", MasqDest:[]string{"256.0.0.0/0"}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidMasqAllowInvalid",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", MasqAllowInvalid:"2"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidMtuFix",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", MtuFix:"2"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidInput",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Input:"FOOL"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidForward",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Forward:"FOOL"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidOutput",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Output:"FOOL"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidFamily",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Family:"FOOL"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidSubnet",
            zone: openwrt.SdewanFirewallZone{Name:"test_zone", Subnet:[]string{"256.0.0.0/0"}},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.CreateZone(tcase.zone)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolZone(t *testing.T) {
    tcases := []struct {
        name string
        zone string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            zone: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            zone: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        err := fw.DeleteZone(tcase.zone)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteAvailableZone(t *testing.T) {
    var err error
    var zone_name = "test_zone_1"
    var redirect_name = "test_redirect_1"
    var zone = openwrt.SdewanFirewallZone{Name:zone_name, Network:[]string{"wan", "lan"}, Input:"REJECT", Masq:"1"}
    var redirect = openwrt.SdewanFirewallRedirect{Name:redirect_name, Src:zone_name, SrcDPort:"22000", Dest:"lan", DestPort:"22", Proto:"tcp", Target:"DNAT"}

    // create zone
    _, err = fw.GetZone(zone_name)
    if (err == nil) {
        err = fw.DeleteZone(zone_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestDeleteAvailableZone: failed to delete zone '%s'", zone_name)
            return
        }
    }

    _, err = fw.CreateZone(zone)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestDeleteAvailableZone: failed to create zone '%s'", zone_name)
        return
    }

    // create redirect
    _, err = fw.GetRedirect(redirect_name)
    if (err == nil) {
        err = fw.DeleteRedirect(redirect_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestDeleteAvailableZone: failed to delete redirect '%s'", redirect_name)
            return
        }
    }

    _, err = fw.CreateRedirect(redirect)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestDeleteAvailableZone: failed to create redirect '%s'", redirect_name)
        return
    }

    // try to delete a used zone
    err = fw.DeleteZone(zone_name)
    if (err != nil) {
        printError(err)
    } else {
        t.Errorf("Test case TestDeleteAvailableZone: error to delete zone '%s'", zone_name)
        return
    }

    // clean
    err = fw.DeleteRedirect(redirect_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestDeleteAvailableZone: failed to delete redirect '%s'", redirect_name)
        return
    }

    err = fw.DeleteZone(zone_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestDeleteAvailableZone: failed to delete zone '%s'", zone_name)
        return
    }
}

func TestUpdateFoolZone(t *testing.T) {
    tcases := []struct {
        name string
        zone openwrt.SdewanFirewallZone
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            zone: openwrt.SdewanFirewallZone{Name:"fool_name"},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.UpdateZone(tcase.zone)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestZone(t *testing.T) {
    var err error
    var ret_zone *openwrt.SdewanFirewallZone

    var zone_name = "test_zone"
    var zone = openwrt.SdewanFirewallZone{Name:zone_name, Network:[]string{"wan", "lan"}, Input:"REJECT", Masq:"1"}
    var update_zone = openwrt.SdewanFirewallZone{Name:zone_name, Network:[]string{"lan"}, Input:"ACCEPT", MasqSrc:[]string{"!0.0.0.0/0", "172.16.1.1"}, Subnet:[]string{"172.16.0.1/24"}}

    _, err = fw.GetZone(zone_name)
    if (err == nil) {
        err = fw.DeleteZone(zone_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestZone: failed to delete zone '%s'", zone_name)
            return
        }
    }

    // Create zone
    ret_zone, err = fw.CreateZone(zone)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestZone: failed to create zone '%s'", zone_name)
        return
    } else {
       p_data, _ := json.Marshal(ret_zone)
        fmt.Printf("Created Zone: %s\n", string(p_data))
    }

    ret_zone, err = fw.GetZone(zone_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestZone: failed to get created zone")
        return
    } else {
        if( ret_zone.Input != "REJECT" ) {
            t.Errorf("Test case TestZone: failed to create zone")
            return
        }
    }

    // Update zone
    ret_zone, err = fw.UpdateZone(update_zone)
    if (err != nil) {
        printError(err)
       t.Errorf("Test case TestZone: failed to update zone '%s'", zone_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_zone)
        fmt.Printf("Updated Zone: %s\n", string(p_data))
    }

    ret_zone, err = fw.GetZone(zone_name)
    if (err != nil) {
        t.Errorf("Test case TestZone: failed to get updated zone")
        return
    } else {
        if( ret_zone.Input != "ACCEPT" ) {
            t.Errorf("Test case TestZone: failed to update zone")
            return
        }
    }

    // Delete zone
    err = fw.DeleteZone(zone_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestZone: failed to delete zone '%s'", zone_name)
        return
    }

    ret_zone, err = fw.GetZone(zone_name)
    if (err == nil) {
        t.Errorf("Test case TestZone: failed to delete zone")
        return
    }
}

// Firewall Forwarding API Test
func TestGetForwardings(t *testing.T) {
    res, err := fw.GetForwardings()
    if res == nil {
        printError(err)
        t.Errorf("Test case GetForwardings: can not get firewall forwardings")
        return
    }

    if len(res.Forwardings) == 0 {
        fmt.Printf("Test case GetForwardings: no forwarding defined")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetForwarding(t *testing.T) {
    tcases := []struct {
        name string
        forwarding string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetFoolForwarding",
            forwarding: "foo_forwarding",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.GetForwarding(tcase.forwarding)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolForwarding(t *testing.T) {
    tcases := []struct {
        name string
        forwarding openwrt.SdewanFirewallForwarding
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            forwarding: openwrt.SdewanFirewallForwarding{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "NoSrc",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"test_forwarding", Dest: available_zone_1},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidSrc",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"test_forwarding", Src:"fool_zone", Dest: available_zone_1},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "NoDest",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"test_forwarding", Src: available_zone_1},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidDest",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"test_forwarding", Src:available_zone_1, Dest:"fool_zone"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidFamily",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"test_forwarding", Family:"fool_family", Src:available_zone_1, Dest:available_zone_2},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.CreateForwarding(tcase.forwarding)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolForwarding(t *testing.T) {
    tcases := []struct {
        name string
        forwarding string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            forwarding: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            forwarding: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        err := fw.DeleteForwarding(tcase.forwarding)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolForwarding(t *testing.T) {
    tcases := []struct {
        name string
        forwarding openwrt.SdewanFirewallForwarding
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            forwarding: openwrt.SdewanFirewallForwarding{Name:"fool_name", Src:available_zone_1, Dest:available_zone_2},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.UpdateForwarding(tcase.forwarding)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestForwarding(t *testing.T) {
    var err error
    var ret_forwarding *openwrt.SdewanFirewallForwarding

    var forwarding_name = "test_forwarding"
    var forwarding = openwrt.SdewanFirewallForwarding{Name:forwarding_name, Src:available_zone_1, Dest:available_zone_2}
    var update_forwarding = openwrt.SdewanFirewallForwarding{Name:forwarding_name, Src:available_zone_2, Dest:available_zone_1}

    _, err = fw.GetForwarding(forwarding_name)
    if (err == nil) {
        err = fw.DeleteForwarding(forwarding_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestForwarding: failed to delete forwarding '%s'", forwarding_name)
            return
        }
    }

    // Create forwarding
    ret_forwarding, err = fw.CreateForwarding(forwarding)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestForwarding: failed to create forwarding '%s'", forwarding_name)
        return
    } else {
       p_data, _ := json.Marshal(ret_forwarding)
        fmt.Printf("Created Forwarding: %s\n", string(p_data))
    }

    ret_forwarding, err = fw.GetForwarding(forwarding_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestForwarding: failed to get created forwarding")
        return
    } else {
        if( ret_forwarding.Dest != available_zone_2 ) {
            t.Errorf("Test case TestForwarding: failed to create forwarding")
            return
        }
    }

    // Update forwarding
    ret_forwarding, err = fw.UpdateForwarding(update_forwarding)
    if (err != nil) {
        printError(err)
       t.Errorf("Test case TestForwarding: failed to update forwarding '%s'", forwarding_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_forwarding)
        fmt.Printf("Updated Forwarding: %s\n", string(p_data))
    }

    ret_forwarding, err = fw.GetForwarding(forwarding_name)
    if (err != nil) {
        t.Errorf("Test case TestForwarding: failed to get updated forwarding")
        return
    } else {
        if( ret_forwarding.Dest != available_zone_1 ) {
            t.Errorf("Test case TestForwarding: failed to update forwarding")
            return
        }
    }

    // Delete forwarding
    err = fw.DeleteForwarding(forwarding_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestForwarding: failed to delete forwarding '%s'", forwarding_name)
        return
    }

    ret_forwarding, err = fw.GetForwarding(forwarding_name)
    if (err == nil) {
        t.Errorf("Test case TestForwarding: failed to delete forwarding")
        return
    }
}

// Firewall Redirect API Test
func TestGetRedirects(t *testing.T) {
    res, err := fw.GetRedirects()
    if res == nil {
        printError(err)
        t.Errorf("Test case GetRedirects: can not get firewall redirects")
        return
    }

    if len(res.Redirects) == 0 {
        fmt.Printf("Test case GetRedirects: no redirect defined")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetRedirect(t *testing.T) {
    tcases := []struct {
        name string
        redirect string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetFoolRedirect",
            redirect: "foo_redirect",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.GetRedirect(tcase.redirect)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolRedirect(t *testing.T) {
    tcases := []struct {
        name string
        redirect openwrt.SdewanFirewallRedirect
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            redirect: openwrt.SdewanFirewallRedirect{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "No Src for DNAT",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Dest:available_zone_2, DestPort:"22", Proto:"tcp", Target:"DNAT"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "No SrcDIp for SNAT",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Dest:available_zone_2, DestPort:"22", Proto:"tcp", Target:"SNAT"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "No Dest for SNAT",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcDIp:"192.168.1.1", DestPort:"22", Proto:"tcp", Target:"SNAT"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Src",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Src:"fool_zone"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcIp",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcIp:"192.168.1.300"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcDIp",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcDIp:"192.168.1.300"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcMac",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcMac:"00:00"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcPort",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcPort:"1a"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcDPort",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", SrcDPort:"1a"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Proto",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Proto:"fool_protocol"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Dest",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Dest:"fool_zone"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid DestIp",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", DestIp:"192.168.1a.1"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid DestPort",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", DestPort:"0"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Target",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Target:"fool_NAT"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Family",
            redirect: openwrt.SdewanFirewallRedirect{Name:"test_redirect", Family:"fool_family"},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.CreateRedirect(tcase.redirect)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolRedirect(t *testing.T) {
    tcases := []struct {
        name string
        redirect string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            redirect: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            redirect: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        err := fw.DeleteRedirect(tcase.redirect)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolRedirect(t *testing.T) {
    tcases := []struct {
        name string
        redirect openwrt.SdewanFirewallRedirect
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            redirect: openwrt.SdewanFirewallRedirect{Name:"fool_name", Src:available_zone_1, SrcDPort:"22000", Dest:available_zone_2, DestPort:"22", Proto:"tcp", Target:"DNAT"},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.UpdateRedirect(tcase.redirect)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestRedirect(t *testing.T) {
    var err error
    var ret_redirect *openwrt.SdewanFirewallRedirect

    var redirect_name = "test_redirect"
    var redirect = openwrt.SdewanFirewallRedirect{Name:redirect_name, Src:available_zone_1, SrcDPort:"22000", Dest:available_zone_2, DestPort:"22", Proto:"tcp", Target:"DNAT"}
    var update_redirect = openwrt.SdewanFirewallRedirect{Name:redirect_name, Src:available_zone_1, SrcDPort:"22001", Dest:available_zone_2, DestPort:"22", Proto:"tcp", Target:"DNAT"}

    _, err = fw.GetRedirect(redirect_name)
    if (err == nil) {
        err = fw.DeleteRedirect(redirect_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestRedirect: failed to delete redirect '%s'", redirect_name)
            return
        }
    }

    // Create redirect
    ret_redirect, err = fw.CreateRedirect(redirect)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRedirect: failed to create redirect '%s'", redirect_name)
        return
    } else {
       p_data, _ := json.Marshal(ret_redirect)
        fmt.Printf("Created Redirect: %s\n", string(p_data))
    }

    ret_redirect, err = fw.GetRedirect(redirect_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRedirect: failed to get created redirect")
        return
    } else {
        if( ret_redirect.SrcDPort != "22000" ) {
            t.Errorf("Test case TestRedirect: failed to create redirect")
            return
        }
    }

    // Update redirect
    ret_redirect, err = fw.UpdateRedirect(update_redirect)
    if (err != nil) {
        printError(err)
       t.Errorf("Test case TestRedirect: failed to update redirect '%s'", redirect_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_redirect)
        fmt.Printf("Updated Redirect: %s\n", string(p_data))
    }

    ret_redirect, err = fw.GetRedirect(redirect_name)
    if (err != nil) {
        t.Errorf("Test case TestRedirect: failed to get updated redirect")
        return
    } else {
        if( ret_redirect.SrcDPort != "22001" ) {
            t.Errorf("Test case TestRedirect: failed to update redirect")
            return
        }
    }

    // Delete redirect
    err = fw.DeleteRedirect(redirect_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRedirect: failed to delete redirect '%s'", redirect_name)
        return
    }

    ret_redirect, err = fw.GetRedirect(redirect_name)
    if (err == nil) {
        t.Errorf("Test case TestRedirect: failed to delete redirect")
        return
    }
}

// Firewall Rule API Test
func TestGetRules(t *testing.T) {
    res, err := fw.GetRules()
    if res == nil {
        printError(err)
        t.Errorf("Test case GetRules: can not get firewall rules")
        return
    }

    if len(res.Rules) == 0 {
        fmt.Printf("Test case GetRules: no rule defined")
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
            name: "GetFoolRule",
            rule: "foo_rule",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.GetRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolRule(t *testing.T) {
    tcases := []struct {
        name string
        rule openwrt.SdewanFirewallRule
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            rule: openwrt.SdewanFirewallRule{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "No SetMark and SetXmark for DNAT",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Target:"MARK"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SRC",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Src:"fool_zone"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcIp",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", SrcIp:"191.11.aa.10"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Valid SrcMac",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", SrcMac:"01:02:a0:0F:aC:EE"},
        },
        {
            name: "Invalid SrcMac",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", SrcMac:"11:rt:00:00:00:00"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid SrcPort",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", SrcPort:"0"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Proto",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Proto:"fool_proto"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid IcmpType",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", IcmpType:[]string{"network-unreachable", "address-mask-reply", "fool_icmp_type"}},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Dest",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Dest:"fool_zone"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid DestIp",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", DestIp:"192.168.1.2.1"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid DestPort",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", DestPort:"aa"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Target",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Target:"fool_target"},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "Invalid Family",
            rule: openwrt.SdewanFirewallRule{Name:"test_rule", Family:"fool_family"},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.CreateRule(tcase.rule)
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
        err := fw.DeleteRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolRule(t *testing.T) {
    tcases := []struct {
        name string
        rule openwrt.SdewanFirewallRule
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            rule: openwrt.SdewanFirewallRule{Name:"fool_name", Src:available_zone_1, Proto:"udp", DestPort:"68", Target:"REJECT", Family:"ipv4"},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

	for _, tcase := range tcases {
        _, err := fw.UpdateRule(tcase.rule)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestRule(t *testing.T) {
    var err error
    var ret_rule *openwrt.SdewanFirewallRule

    var rule_name = "test_rule"
    var rule = openwrt.SdewanFirewallRule{Name:rule_name, Src:available_zone_1, Proto:"udp", DestPort:"68", Target:"REJECT", Family:"ipv4"}
    var update_rule = openwrt.SdewanFirewallRule{Name:rule_name, Src:available_zone_1, IcmpType:[]string{"host-redirect", "echo-request"}, Proto:"udp", DestPort:"68", Target:"ACCEPT", Family:"ipv4"}

    _, err = fw.GetRule(rule_name)
    if (err == nil) {
        err = fw.DeleteRule(rule_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestRule: failed to delete rule '%s'", rule_name)
            return
        }
    }

    // Create rule
    ret_rule, err = fw.CreateRule(rule)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to create rule '%s'", rule_name)
        return
    } else {
       p_data, _ := json.Marshal(ret_rule)
        fmt.Printf("Created Rule: %s\n", string(p_data))
    }

    ret_rule, err = fw.GetRule(rule_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to get created rule")
        return
    } else {
        if( ret_rule.Target != "REJECT" ) {
            t.Errorf("Test case TestRule: failed to create rule")
            return
        }
    }

    // Update rule
    ret_rule, err = fw.UpdateRule(update_rule)
    if (err != nil) {
        printError(err)
       t.Errorf("Test case TestRule: failed to update rule '%s'", rule_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_rule)
        fmt.Printf("Updated Rule: %s\n", string(p_data))
    }

    ret_rule, err = fw.GetRule(rule_name)
    if (err != nil) {
        t.Errorf("Test case TestRule: failed to get updated rule")
        return
    } else {
        if( ret_rule.Target != "ACCEPT" ) {
            t.Errorf("Test case TestRule: failed to update rule")
            return
        }
    }

    // Delete rule
    err = fw.DeleteRule(rule_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestRule: failed to delete rule '%s'", rule_name)
        return
    }

    ret_rule, err = fw.GetRule(rule_name)
    if (err == nil) {
        t.Errorf("Test case TestRule: failed to delete rule")
        return
    }
}
