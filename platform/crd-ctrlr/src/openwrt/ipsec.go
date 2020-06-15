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
	Name                string `json:"name"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	HashAlgorithm       string `json:"hash_algorithm"`
	DhGroup             string `json:"dh_group"`
}

type SdewanIpsecProposals struct {
	Proposals []SdewanIpsecProposal `json:"proposals"`
}

func (o *SdewanIpsecProposal) GetName() string {
	return o.Name
}

// Remotes
type SdewanIpsecConnection struct {
	Name           string   `json:"name"`
	ConnType       string   `json:"conn_type"`
	Mode           string   `json:"mode"`
	LocalSubnet    string   `json:"local_subnet"`
	LocalNat       string   `json:"local_nat"`
	LocalSourceip  string   `json:"local_sourceip"`
	LocalUpdown    string   `json:"local_updown"`
	LocalFirewall  string   `json:"local_firewall"`
	RemoteSubnet   string   `json:"remote_subnet"`
	RemoteSourceip string   `json:"remote_sourceip"`
	RemoteUpdown   string   `json:"remote_updown"`
	RemoteFirewall string   `json:"remote_firewall"`
	CryptoProposal []string `json:"crypto_proposal"`
	Mark           string   `json:"mark"`
	IfId           string   `json:"if_id"`
}

type SdewanIpsecRemote struct {
	Name                 string                  `json:"name"`
	Type                 string                  `json:"type"`
	Gateway              string                  `json:"gateway"`
	PreSharedKey         string                  `json:"pre_shared_key"`
	AuthenticationMethod string                  `json:"authentication_method"`
	LocalIdentifier      string                  `json:"local_identifier"`
	RemoteIdentifier     string                  `json:"remote_identifier"`
	CryptoProposal       []string                `json:"crypto_proposal"`
	ForceCryptoProposal  string                  `json:"force_crypto_proposal"`
	LocalPublicCert      string                  `json:"local_public_cert"`
	LocalPrivateCert     string                  `json:"local_private_cert"`
	SharedCa             string                  `json:"shared_ca"`
	Connections          []SdewanIpsecConnection `json:"connections"`
}

type SdewanIpsecRemotes struct {
	Remotes []SdewanIpsecRemote `json:"remotes"`
}

func (o *SdewanIpsecRemote) GetName() string {
	return o.Name
}

// Proposal APIs
// get proposals
func (f *IpsecClient) GetProposals() (*SdewanIpsecProposals, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(ipsecBaseURL + "proposals")
	if err != nil {
		return nil, err
	}

	var sdewanIpsecProposals SdewanIpsecProposals
	err = json.Unmarshal([]byte(response), &sdewanIpsecProposals)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecProposals, nil
}

// get proposal
func (m *IpsecClient) GetProposal(proposal string) (*SdewanIpsecProposal, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(ipsecBaseURL + "proposals/" + proposal)
	if err != nil {
		return nil, err
	}

	var sdewanIpsecProposal SdewanIpsecProposal
	err = json.Unmarshal([]byte(response), &sdewanIpsecProposal)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecProposal, nil
}

// create proposal
func (m *IpsecClient) CreateProposal(proposal SdewanIpsecProposal) (*SdewanIpsecProposal, error) {
	var response string
	var err error
	proposal_obj, _ := json.Marshal(proposal)
	response, err = m.OpenwrtClient.Post(ipsecBaseURL+"proposals", string(proposal_obj))
	if err != nil {
		return nil, err
	}

	var sdewanIpsecProposal SdewanIpsecProposal
	err = json.Unmarshal([]byte(response), &sdewanIpsecProposal)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecProposal, nil
}

// delete proposal
func (m *IpsecClient) DeleteProposal(proposal_name string) error {
	_, err := m.OpenwrtClient.Delete(ipsecBaseURL + "proposals/" + proposal_name)
	if err != nil {
		return err
	}

	return nil
}

// update proposal
func (m *IpsecClient) UpdateProposal(proposal SdewanIpsecProposal) (*SdewanIpsecProposal, error) {
	var response string
	var err error
	proposal_obj, _ := json.Marshal(proposal)
	proposal_name := proposal.Name
	response, err = m.OpenwrtClient.Put(ipsecBaseURL+"proposals/"+proposal_name, string(proposal_obj))
	if err != nil {
		return nil, err
	}

	var sdewanIpsecProposal SdewanIpsecProposal
	err = json.Unmarshal([]byte(response), &sdewanIpsecProposal)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecProposal, nil
}

// Remote APIs
// get remotes
func (f *IpsecClient) GetRemotes() (*SdewanIpsecRemotes, error) {
	var response string
	var err error
	response, err = f.OpenwrtClient.Get(ipsecBaseURL + "remotes")
	if err != nil {
		return nil, err
	}

	var sdewanIpsecRemotes SdewanIpsecRemotes
	err = json.Unmarshal([]byte(response), &sdewanIpsecRemotes)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecRemotes, nil
}

// get remote
func (m *IpsecClient) GetRemote(remote string) (*SdewanIpsecRemote, error) {
	var response string
	var err error
	response, err = m.OpenwrtClient.Get(ipsecBaseURL + "remotes/" + remote)
	if err != nil {
		return nil, err
	}

	var sdewanIpsecRemote SdewanIpsecRemote
	err = json.Unmarshal([]byte(response), &sdewanIpsecRemote)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecRemote, nil
}

// create remote
func (m *IpsecClient) CreateRemote(remote SdewanIpsecRemote) (*SdewanIpsecRemote, error) {
	var response string
	var err error
	remote_obj, _ := json.Marshal(remote)
	response, err = m.OpenwrtClient.Post(ipsecBaseURL+"remotes", string(remote_obj))
	if err != nil {
		return nil, err
	}

	var sdewanIpsecRemote SdewanIpsecRemote
	err = json.Unmarshal([]byte(response), &sdewanIpsecRemote)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecRemote, nil
}

// delete remote
func (m *IpsecClient) DeleteRemote(remote_name string) error {
	_, err := m.OpenwrtClient.Delete(ipsecBaseURL + "remotes/" + remote_name)
	if err != nil {
		return err
	}

	return nil
}

// update remote
func (m *IpsecClient) UpdateRemote(remote SdewanIpsecRemote) (*SdewanIpsecRemote, error) {
	var response string
	var err error
	remote_obj, _ := json.Marshal(remote)
	remote_name := remote.Name
	response, err = m.OpenwrtClient.Put(ipsecBaseURL+"remotes/"+remote_name, string(remote_obj))
	if err != nil {
		return nil, err
	}

	var sdewanIpsecRemote SdewanIpsecRemote
	err = json.Unmarshal([]byte(response), &sdewanIpsecRemote)
	if err != nil {
		return nil, err
	}

	return &sdewanIpsecRemote, nil
}
