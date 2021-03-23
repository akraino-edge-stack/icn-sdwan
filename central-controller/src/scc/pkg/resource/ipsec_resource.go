/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governinog permissions and
 * limitations under the License.
 */

package resource

import (
        "strings"
)

const (
    AuthTypePSK = "psk"
    AuthTypePUBKEY = "pubkey"
)

type Connection struct {
        Name           string
        ConnectionType string
        Mode           string
        LocalSubnet    string
        LocalSourceIp  string
        LocalUpDown    string
        LocalFirewall  string
        RemoteSubnet   string
        RemoteSourceIp string
        RemoteUpDown   string
        RemoteFirewall string
        CryptoProposal []string
        Mark           string
        IfId           string
}

type IpsecResource struct {
        Name                 string
        Type                 string
        Remote               string
        AuthenticationMethod string
        CryptoProposal       []string
        LocalIdentifier      string
        RemoteIdentifier     string
        ForceCryptoProposal  string
        PresharedKey         string
        PublicCert      string
        PrivateCert     string
        SharedCA             string
        Connections          Connection
}

func (c *IpsecResource) GetName() string {
        return c.Name
}

func (c *IpsecResource) GetType() string {
        return "Ipsec"
}

func (c *IpsecResource) ToYaml() string {
        p := strings.Join(c.CryptoProposal, ",")
        pr := strings.Join(c.Connections.CryptoProposal, ",")
	var connection = ""

        if c.Connections.LocalSubnet != "" {
          base := `apiVersion: ` + SdewanApiVersion + ` 
kind: IpsecSite
metadata:
  name: ` +  c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
spec:
  type: ` + c.Type + `
  remote: '` + c.Remote + `'
  authentication_method: `+ c.AuthenticationMethod + `
  force_crypto_proposal: "` + c.ForceCryptoProposal + `
  crypto_proposal: [` + p + `]`

          connection = `
  connections: 
  - name: ` + c.Connections.Name + `
    conn_type: ` + c.Connections.ConnectionType + `
    mode: ` +  c.Connections.Mode + `
    mark: "` +  c.Connections.Mark + `"
    local_updown: ` + c.Connections.LocalUpDown + `
    local_subnet: ` + c.Connections.LocalSubnet + `
    crypto_proposal: [` + pr +`]`

          if c.Connections.RemoteSourceIp != "" {
            remote_source_ip := `
    remote_source_ip: '` + c.Connections.RemoteSourceIp + `'`
            connection += remote_source_ip
          }

          if c.AuthenticationMethod == AuthTypePUBKEY {
            auth := `
  local_public_cert: ` + c.PublicCert + `
  local_private_cert: ` + c.PrivateCert + `
  shared_ca: ` + c.SharedCA + `
  local_identifier: ` + c.LocalIdentifier + `
  remote_identifier: ` + c.RemoteIdentifier
            return base + auth + connection
          } else if c.AuthenticationMethod == AuthTypePSK {
            auth := `
  pre_shared_key: ` + c.PresharedKey + `
  local_identifier: ` + c.LocalIdentifier + `
  remote_identifier: ` + c.RemoteIdentifier
            return base + auth + connection
          } else {
            return "Error in authentication method"
          }

        }

        base := `apiVersion: ` + SdewanApiVersion + ` 
kind: IpsecHost
metadata:
  name: ` +  c.Name + `
  namespace: default
  labels:
    sdewanPurpose: ` + SdewanPurpose + `
spec:
  type: ` + c.Type + `
  remote: '` + c.Remote + `'
  authentication_method: `+ c.AuthenticationMethod +`
  force_crypto_proposal: "` + c.ForceCryptoProposal + `"
  crypto_proposal: [` + p + `]`

        if c.Connections.LocalSourceIp != "" {
          connection = `
  connections: 
  - name: ` + c.Connections.Name + `
    conn_type: ` + c.Connections.ConnectionType + `
    mode: ` +  c.Connections.Mode + `
    mark: "` +  c.Connections.Mark + `"
    local_updown: ` + c.Connections.LocalUpDown + `
    local_sourceip: '` + c.Connections.LocalSourceIp + `'
    crypto_proposal: [` + pr +`]`
        } else {
          connection = `
  connections: 
  - name: ` + c.Connections.Name + `
    conn_type: ` + c.Connections.ConnectionType + `
    mode: ` +  c.Connections.Mode + `
    mark: "` +  c.Connections.Mark + `"
    local_updown: ` + c.Connections.LocalUpDown + `
    crypto_proposal: [` + pr +`]`
        }

        if c.Connections.RemoteSourceIp != "" {
          remote_source_ip := `
    remote_sourceip: '` + c.Connections.RemoteSourceIp + `'`
          connection += remote_source_ip
        }

        if c.AuthenticationMethod == AuthTypePUBKEY {
          auth := `
  local_public_cert: ` + c.PublicCert + `
  local_private_cert: ` + c.PrivateCert + `
  shared_ca: ` + c.SharedCA + `
  local_identifier: ` + c.LocalIdentifier + `
  remote_identifier: ` + c.RemoteIdentifier
          return base + auth + connection
        } else if c.AuthenticationMethod == AuthTypePSK {
          auth := `
  pre_shared_key: ` + c.PresharedKey + `
  local_identifier: ` + c.LocalIdentifier + `
  remote_identifier: ` + c.RemoteIdentifier
          return base + auth + connection
        } else {
          return "Error in authentication method"
        }

}

func init() {
  GetResourceBuilder().Register("Ipsec", &IpsecResource{})
}
