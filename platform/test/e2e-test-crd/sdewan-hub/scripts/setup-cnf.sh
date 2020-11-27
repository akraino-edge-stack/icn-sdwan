#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

base=$(pwd)

test -f $base/variables
. $base/variables
providerSubnet=${providerSubnet}
providerGateway=${providerGateway}
providerExcludeIps=${providerExcludeIps}
providerNetworkInterface=${providerNetworkInterface}
cnfWanGateway=${cnfWanGateway}

clean()
{
echo "Cleaning ..."
kubectl delete -f network-prepare.yaml
kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
[-f ipsec_config.yaml ] && kubectl delete -f ipsec_config.yaml
[-f ipsec_proposal.yaml ] && kubectl delete -f ipsec_proposal.yaml
}

error_detect()
{
	echo "Error on line $1"
	#clean
}

trap "error_detect $LINENO" ERR

echo "--------------------- Setup CNF for sdewan hub -----------------------"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml --validate=false
sleep 2m

echo "--------------------- Creating ovn networks ---------------------"
cat > network-prepare.yaml << EOF
---
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: ProviderNetwork
metadata:
  name: pnetwork
spec:
  cniType: ovn4nfv
  ipv4Subnets:
  - subnet: $providerSubnet
    name: subnet
    gateway: $providerGateway
    excludeIps: $providerExcludeIps
  providerNetType: DIRECT
  direct:
    providerInterfaceName: $providerNetworkInterface
    directNodeSelector: all

EOF

kubectl apply -f network-prepare.yaml
sleep 2

ovnProviderNet=$(kubectl get providernetwork | sed -n 2p | awk '{print $1}')
if [ -n "${ovnProviderNet}" ]
then
	echo "Network created successfully"
else
	echo "Network creation failed"
	exit 1
fi


echo "--------------------- Creating sdwan-cnf with helm ---------------------"
curl https://helm.baltorepo.com/organization/signing.asc | sudo apt-key add -
sudo apt-get install apt-transport-https --yes
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm

envsubst < ./cnf/values.yaml >> ./cnf/values.yaml
helm package ./cnf
helm install ./cnf-0.1.0.tgz --generate-name

sleep 20

sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}' | head -1)
if [ "$sdwan_status" == "Running" ]
then
	echo "Sdewan cnf $sdewan_cnf_name created successfully"
else
        sleep 40
	sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}' | head -1)
	if [ "$sdwan_status" != "Running" ]
	then
	     echo "Sdewan cnf creation failed"
             exit 2
        fi
fi

echo "--------------------- Setup sdewan controller ---------------------"
helm package ./controllers
helm install ./controllers-0.1.0.tgz --generate-name
sleep 1m

echo "--------------------- Applying CRDs ---------------------"
cat > ipsec_proposal.yaml << EOF
---
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: IpsecProposal
metadata:
  name: ipsecproposal
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
  dh_group: modp3072
  encryption_algorithm: aes128
  hash_algorithm: sha256

EOF

kubectl apply -f ipsec_proposal.yaml

cat > ipsec_config.yaml << EOF
---
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: IpsecSite
metadata:
  name: ipsecsite
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
    name: sdewan-hub 
    remote: "%any" 
    pre_shared_key: test_key
    authentication_method: psk
    local_identifier: $hubIp
    crypto_proposal:
      - ipsecproposal
    force_crypto_proposal: "0"
    connections:
    - name: connA
      conn_type: tunnel
      mode: start
      remote_sourceip: "192.168.1.5-192.168.1.6"
      local_subnet: 192.168.1.1/24,$hubIp/32
      crypto_proposal:
        - ipsecproposal

EOF

kubectl apply -f ipsec_config.yaml

echo "--------------------- Configuration finished ---------------------"
