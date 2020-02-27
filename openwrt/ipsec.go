package openwrt

import (
    "encoding/json"
)

const (
    ipsecBaseURL = "sdewan/ipsec/v1/"
)

type IpsecClient struct {
    OpenwrtClient *openwrtClient
}

// Proposals
type SdewanIpsecProposal struct {
    Name string `json:"name"`
    EncryptionAlgorithm string `json:"encryption_algorithm"`
    HashAlgorithm string `json:"hash_algorithm"`
    DhGroup string `json:"dh_group"`
}

type SdewanIpsecProposals struct {
    Proposals []SdewanIpsecProposal `json:"proposals"`
}

// Sites
type SdewanIpsecConnection struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Mode string `json:"mode"`
    LocalSubnet string `json:"local_subnet"`
    LocalNat string `json:"local_nat"`
    LocalSourceip string `json:"local_sourceip"`
    LocalUpdown string `json:"local_updown"`
    LocalFirewall string `json:"local_firewall"`
    RemoteSubnet string `json:"remote_subnet"`
    RemoteSourceip string `json:"remote_sourceip"`
    RemoteUpdown string `json:"remote_updown"`
    RemoteFirewall string `json:"remote_firewall"`
    CryptoProposal []string `json:"crypto_proposal"`
}

type SdewanIpsecSite struct {
    Name string `json:"name"`
    Gateway string `json:"gateway"`
    PreSharedKey string `json:"pre_shared_key"`
    AuthenticationMethod string `json:"authentication_method"`
    LocalIdentifier string `json:"local_identifier"`
    RemoteIdentifier string `json:"remote_identifier"`
    CryptoProposal []string `json:"crypto_proposal"`
    ForceCryptoProposal string `json:"force_crypto_proposal"`
    LocalPublicCert string `json:"local_public_cert"`
    LocalPrivateCert string `json:"local_private_cert"`
    SharedCa string `json:"shared_ca"`
    Connections []SdewanIpsecConnection `json:"connections"`
}

type SdewanIpsecSites struct {
    Sites []SdewanIpsecSite `json:"sites"`
}

// Proposal APIs
// get proposals
func (f *IpsecClient) GetProposals() (*SdewanIpsecProposals, error) {
    response, err := f.OpenwrtClient.Get(ipsecBaseURL + "proposals")
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecProposals SdewanIpsecProposals
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecProposals)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecProposals, nil
}

// get proposal
func (m *IpsecClient) GetProposal(proposal string) (*SdewanIpsecProposal, error) {
    response, err := m.OpenwrtClient.Get(ipsecBaseURL + "proposal/" + proposal)
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecProposal SdewanIpsecProposal
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecProposal)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecProposal, nil
}

// create proposal
func (m *IpsecClient) CreateProposal(proposal SdewanIpsecProposal) (*SdewanIpsecProposal, error) {
    proposal_obj, _ := json.Marshal(proposal)
    response, err := m.OpenwrtClient.Post(ipsecBaseURL + "proposal", string(proposal_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecProposal SdewanIpsecProposal
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecProposal)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecProposal, nil
}

// delete proposal
func (m *IpsecClient) DeleteProposal(proposal_name string) (error) {
    _, err := m.OpenwrtClient.Delete(ipsecBaseURL + "proposal/" + proposal_name)
    if (err != nil) {
        return err
    }

    return nil
}

// update proposal
func (m *IpsecClient) UpdateProposal(proposal SdewanIpsecProposal) (*SdewanIpsecProposal, error) {
    proposal_obj, _ := json.Marshal(proposal)
    proposal_name := proposal.Name
    response, err := m.OpenwrtClient.Put(ipsecBaseURL + "proposal/" + proposal_name, string(proposal_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecProposal SdewanIpsecProposal
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecProposal)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecProposal, nil
}

// Site APIs
// get sites
func (f *IpsecClient) GetSites() (*SdewanIpsecSites, error) {
    response, err := f.OpenwrtClient.Get(ipsecBaseURL + "sites")
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecSites SdewanIpsecSites
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecSites)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecSites, nil
}

// get site
func (m *IpsecClient) GetSite(site string) (*SdewanIpsecSite, error) {
    response, err := m.OpenwrtClient.Get(ipsecBaseURL + "site/" + site)
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecSite SdewanIpsecSite
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecSite)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecSite, nil
}

// create site
func (m *IpsecClient) CreateSite(site SdewanIpsecSite) (*SdewanIpsecSite, error) {
    site_obj, _ := json.Marshal(site)
    response, err := m.OpenwrtClient.Post(ipsecBaseURL + "site", string(site_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecSite SdewanIpsecSite
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecSite)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecSite, nil
}

// delete site
func (m *IpsecClient) DeleteSite(site_name string) (error) {
    _, err := m.OpenwrtClient.Delete(ipsecBaseURL + "site/" + site_name)
    if (err != nil) {
        return err
    }

    return nil
}

// update site
func (m *IpsecClient) UpdateSite(site SdewanIpsecSite) (*SdewanIpsecSite, error) {
    site_obj, _ := json.Marshal(site)
    site_name := site.Name
    response, err := m.OpenwrtClient.Put(ipsecBaseURL + "site/" + site_name, string(site_obj))
    if (err != nil) {
        return nil, err
    }

    var sdewanIpsecSite SdewanIpsecSite
    err2 := json.Unmarshal([]byte(response), &sdewanIpsecSite)
    if (err2 != nil) {
        return nil, err2
    }

    return &sdewanIpsecSite, nil
}
