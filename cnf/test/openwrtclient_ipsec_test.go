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

var ic openwrt.IpsecClient
var available_proposal_1 string
var available_proposal_2 string
var available_site string

func TestMain(m *testing.M) {
    servIp := flag.String("ip", "10.244.0.18", "SDEWAN CNF Management IP Address")
    flag.Parse()

    var err error
    client := openwrt.NewOpenwrtClient(*servIp, "root", "")
    ic = openwrt.IpsecClient{client}
    available_proposal_1 = "test_proposal1"
    available_proposal_2 = "test_proposal2"
    available_site = "test_default_site"

    // create default proposals
    var proposall = openwrt.SdewanIpsecProposal{Name:available_proposal_1, EncryptionAlgorithm:"aes128", HashAlgorithm:"sha256", DhGroup:"modp3072",}
    var proposal2 = openwrt.SdewanIpsecProposal{Name:available_proposal_2, EncryptionAlgorithm:"aes256", HashAlgorithm:"sha128", DhGroup:"modp4096",}

    _, err = ic.GetProposal(available_proposal_1)
    if (err != nil) {
        // Create proposal
        _, err = ic.CreateProposal(proposall)
        if (err != nil) {
            printError(err)
            return
        }
    }

    _, err = ic.GetProposal(available_proposal_2)
    if (err != nil) {
        // Create proposal
        _, err = ic.CreateProposal(proposal2)
        if (err != nil) {
            printError(err)
            return
        }
    }

    // create default site
    var site = openwrt.SdewanIpsecSite{
        Name:available_site,
        Gateway:"10.0.1.2",
        PreSharedKey:"test_key",
        AuthenticationMethod:"psk",
        LocalIdentifier:"C=CH, O=strongSwan, CN=peer",
        RemoteIdentifier:"C=CH, O=strongSwan, CN=peerB",
        ForceCryptoProposal:"true",
        LocalPublicCert:"public cert\npublic cert value",
        CryptoProposal:[]string{available_proposal_1},
        Connections:[]openwrt.SdewanIpsecConnection{{
                Name:available_site+"_conn",
                Type:"tunnel",
                Mode:"start",
                LocalSubnet:"192.168.1.1/24",
                LocalSourceip:"10.0.1.1",
                RemoteSubnet:"192.168.0.1/24",
                RemoteSourceip:"10.0.1.2",
                CryptoProposal:[]string{available_proposal_1, available_proposal_2},
            },
        },
    }

    _, err = ic.GetSite(available_site)
    if (err == nil) {
        // Update site
        _, err = ic.UpdateSite(site)
        if (err != nil) {
            printError(err)
            return
        }
    } else {
        // Create site
        _, err = ic.CreateSite(site)
        if (err != nil) {
            printError(err)
            return
        }
    }

    var ret = m.Run()

    // clean
    ic.DeleteSite(available_site)
    ic.DeleteProposal(available_proposal_1)
    ic.DeleteProposal(available_proposal_2)

    os.Exit(ret)
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

// IpSec Site API Test
func TestGetSites(t *testing.T) {
    res, err := ic.GetSites()
    if res == nil {
        printError(err)
        t.Errorf("Test case GetSites: can not get IpSec sites")
        return
    }

    if len(res.Sites) == 0 {
        fmt.Printf("Test case GetSites: no site found")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetSite(t *testing.T) {
    tcases := []struct {
        name string
        site string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetAvailableSite",
            site: available_site,
        },
        {
            name: "GetFoolSite",
            site: "foo_site",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.GetSite(tcase.site)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolSite(t *testing.T) {
    tcases := []struct {
        name string
        site openwrt.SdewanIpsecSite
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            site: openwrt.SdewanIpsecSite{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidConnectionType",
            site: openwrt.SdewanIpsecSite{Name:"test_site",
                Connections:[]openwrt.SdewanIpsecConnection{{Name:"test_connection_name", Type:"fool_type",},},},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidCryptoProtocolInConnection",
            site: openwrt.SdewanIpsecSite{Name:"test_site",
                Connections:[]openwrt.SdewanIpsecConnection{{Name:"test_connection_name", Type:"tunnel", CryptoProposal:[]string{available_proposal_1, "fool_protocol"}},},},
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "InvalidCryptoProtocolInSite",
            site: openwrt.SdewanIpsecSite{Name:"test_site", CryptoProposal:[]string{available_proposal_1, "fool_protocol"},},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.CreateSite(tcase.site)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolSite(t *testing.T) {
    tcases := []struct {
        name string
        site string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            site: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            site: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        err := ic.DeleteSite(tcase.site)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolSite(t *testing.T) {
    tcases := []struct {
        name string
        site openwrt.SdewanIpsecSite
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            site: openwrt.SdewanIpsecSite{Name:"fool_name"},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.UpdateSite(tcase.site)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestSite(t *testing.T) {
    var err error
    var ret_site *openwrt.SdewanIpsecSite

    var site_name = "test_site"
    var conn_name1 = "test_name1"
    var conn_name2 = "test_name2"
    var conn_name3 = "test_name3"
    var site = openwrt.SdewanIpsecSite{
        Name:site_name,
        Gateway:"10.0.1.2",
        PreSharedKey:"test_key",
        AuthenticationMethod:"psk",
        LocalIdentifier:"C=CH, O=strongSwan, CN=peer",
        RemoteIdentifier:"C=CH, O=strongSwan, CN=peerB",
        ForceCryptoProposal:"true",
        CryptoProposal:[]string{available_proposal_1},
        Connections:[]openwrt.SdewanIpsecConnection{{
                Name:conn_name1,
                Type:"tunnel",
                Mode:"start",
                LocalSubnet:"192.168.1.1/24",
                LocalSourceip:"10.0.1.1",
                RemoteSubnet:"192.168.0.1/24",
                RemoteSourceip:"10.0.1.2",
                CryptoProposal:[]string{available_proposal_1, available_proposal_2},
            },
        },
    }

    var update_site = openwrt.SdewanIpsecSite{
        Name:site_name,
        Gateway:"10.0.21.2",
        PreSharedKey:"test_key_2",
        AuthenticationMethod:"psk",
        LocalIdentifier:"C=CH, O=strongSwan, CN=peer",
        RemoteIdentifier:"C=CH, O=strongSwan, CN=peerB",
        ForceCryptoProposal:"true",
        CryptoProposal:[]string{available_proposal_1, available_proposal_2},
        Connections:[]openwrt.SdewanIpsecConnection{{
                Name:conn_name1,
                Type:"tunnel",
                Mode:"start",
                LocalSubnet:"192.168.21.1/24",
                LocalSourceip:"10.0.1.1",
                RemoteSubnet:"192.168.0.1/24",
                RemoteSourceip:"10.0.21.2",
                CryptoProposal:[]string{available_proposal_2},
            },
            {
                Name:conn_name2,
                Type:"transport",
                Mode:"start",
                LocalSubnet:"192.168.31.1/24",
                LocalSourceip:"10.0.11.1",
                RemoteSubnet:"192.168.10.1/24",
                RemoteSourceip:"10.0.31.2",
                CryptoProposal:[]string{available_proposal_1, available_proposal_2},
            },
            {
                Name:conn_name3,
                Type:"tunnel",
                Mode:"start",
                LocalSubnet:"192.168.41.1/24",
                LocalSourceip:"10.0.11.1",
                RemoteSubnet:"192.168.10.1/24",
                RemoteSourceip:"10.0.41.2",
                CryptoProposal:[]string{available_proposal_1},
            },
        },
    }

    _, err = ic.GetSite(site_name)
    if (err == nil) {
        err = ic.DeleteSite(site_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestSite: failed to delete site '%s'", site_name)
            return
        }
    }

    // Create site
    ret_site, err = ic.CreateSite(site)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestSite: failed to create site '%s'", site_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_site)
        fmt.Printf("Created Site: %s\n", string(p_data))
    }

    ret_site, err = ic.GetSite(site_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestSite: failed to get created site")
        return
    } else {
        if( len(ret_site.Connections) != 1 || ret_site.Connections[0].LocalSubnet != "192.168.1.1/24" ) {
            t.Errorf("Test case TestSite: failed to create site")
            return
        }
    }

    // Update site
    ret_site, err = ic.UpdateSite(update_site)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestSite: failed to update site '%s'", site_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_site)
        fmt.Printf("Updated Site: %s\n", string(p_data))
    }

    ret_site, err = ic.GetSite(site_name)
    if (err != nil) {
        t.Errorf("Test case TestSite: failed to get updated site")
        return
    } else {
        if( len(ret_site.Connections) != 3 || ret_site.Connections[0].LocalSubnet != "192.168.21.1/24" ) {
            t.Errorf("Test case TestSite: failed to update site")
            return
        }
    }

    // Delete site
    err = ic.DeleteSite(site_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestSite: failed to delete site '%s'", site_name)
        return
    }

    ret_site, err = ic.GetSite(site_name)
    if (err == nil) {
        t.Errorf("Test case TestSite: failed to delete site")
        return
    }
}

// IpSec Proposal API Test
func TestGetProposals(t *testing.T) {
    res, err := ic.GetProposals()
    if res == nil {
        printError(err)
        t.Errorf("Test case TestGetProposals: can not get IpSec proposals")
        return
    }

    if len(res.Proposals) == 0 {
        fmt.Printf("Test case TestGetProposals: no proposal found")
        return
    }

    p_data, _ := json.Marshal(res)
    fmt.Printf("%s\n", string(p_data))
}

func TestGetProposal(t *testing.T) {
    tcases := []struct {
        name string
        proposal string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "GetAvailableProposal",
            proposal: available_proposal_1,
        },
        {
            name: "GetFoolProposal",
            proposal: "foo_proposal",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.GetProposal(tcase.proposal)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateFoolProposal(t *testing.T) {
    tcases := []struct {
        name string
        proposal openwrt.SdewanIpsecProposal
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            proposal: openwrt.SdewanIpsecProposal{Name:" "},
            expectedErr: true,
            expectedErrCode: 400,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.CreateProposal(tcase.proposal)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteFoolProposal(t *testing.T) {
    tcases := []struct {
        name string
        proposal string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            proposal: "",
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "DeleteInUsedProposal",
            proposal: available_proposal_1,
            expectedErr: true,
            expectedErrCode: 400,
        },
        {
            name: "FoolName",
            proposal: "fool_name",
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        err := ic.DeleteProposal(tcase.proposal)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateFoolProposal(t *testing.T) {
    tcases := []struct {
        name string
        proposal openwrt.SdewanIpsecProposal
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            proposal: openwrt.SdewanIpsecProposal{Name:"fool_name"},
            expectedErr: true,
            expectedErrCode: 404,
        },
    }

    for _, tcase := range tcases {
        _, err := ic.UpdateProposal(tcase.proposal)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestProposal(t *testing.T) {
    var err error
    var ret_proposal *openwrt.SdewanIpsecProposal

    var proposal_name = "test_proposal"
    var proposal = openwrt.SdewanIpsecProposal{Name:proposal_name,EncryptionAlgorithm:"aes128", HashAlgorithm:"sha256", DhGroup:"modp3072",}
    var update_proposal = openwrt.SdewanIpsecProposal{Name:proposal_name,EncryptionAlgorithm:"aes256", HashAlgorithm:"sha128", DhGroup:"modp3072",}

    _, err = ic.GetProposal(proposal_name)
    if (err == nil) {
        err = ic.DeleteProposal(proposal_name)
        if (err != nil) {
            printError(err)
            t.Errorf("Test case TestProposal: failed to delete proposal '%s'", proposal_name)
            return
        }
    }

    // Create proposal
    ret_proposal, err = ic.CreateProposal(proposal)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestProposal: failed to create proposal '%s'", proposal_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_proposal)
        fmt.Printf("Created Proposal: %s\n", string(p_data))
    }

    ret_proposal, err = ic.GetProposal(proposal_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestProposal: failed to get created proposal")
        return
    } else {
        if( ret_proposal.EncryptionAlgorithm != "aes128" ) {
            t.Errorf("Test case TestProposal: failed to create proposal")
            return
        }
    }

    // Update proposal
    ret_proposal, err = ic.UpdateProposal(update_proposal)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestProposal: failed to update proposal '%s'", proposal_name)
        return
    } else {
        p_data, _ := json.Marshal(ret_proposal)
        fmt.Printf("Updated Proposal: %s\n", string(p_data))
    }

    ret_proposal, err = ic.GetProposal(proposal_name)
    if (err != nil) {
        t.Errorf("Test case TestProposal: failed to get updated proposal")
        return
    } else {
        if( ret_proposal.EncryptionAlgorithm != "aes256" ) {
            t.Errorf("Test case TestProposal: failed to update proposal")
            return
        }
    }

    // Delete proposal
    err = ic.DeleteProposal(proposal_name)
    if (err != nil) {
        printError(err)
        t.Errorf("Test case TestProposal: failed to delete proposal '%s'", proposal_name)
        return
    }

    ret_proposal, err = ic.GetProposal(proposal_name)
    if (err == nil) {
        t.Errorf("Test case TestProposal: failed to delete proposal")
        return
    }
}
