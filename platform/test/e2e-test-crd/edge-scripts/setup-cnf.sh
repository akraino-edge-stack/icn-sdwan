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
ovnSubnet=${ovnSubnet}
ovnGateway=${ovnGateway}
ovnExcludeIps=${ovnExcludeIps}
cnfLan0=${cnfLan0}
appLan0=${appLan0}
cnfWanGateway=${cnfWanGateway}
app_pod_name=${app_pod_name}

clean()
{
echo "Cleaning ..."
kubectl delete -f httpbin-svc.yaml
kubectl delete -f network-prepare.yaml
kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
[ -f firewall_snat.yaml ] && kubectl delete -f firewall_snat.yaml
[ -f firewall_zone.yaml ] && kubectl delete -f firewall_zone.yaml
[ -f ipsec_config.yaml ] && kubectl delete -f ipsec_config.yaml
[ -f ipsec_proposal.yaml ] && kubectl delete -f ipsec_proposal.yaml
helm delete $(helm list -q)
}

error_detect()
{
	echo "Error on line $1"
	clean
}

trap "error_detect $LINENO" ERR

echo "--------------------- Setup CNF for ${edgeName} ---------------------"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml --validate=false
sleep 3m

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

---
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: ovn-network
spec:
  # Add fields here
  cniType: ovn4nfv
  ipv4Subnets:
  - subnet: $ovnSubnet
    name: subnet1
    gateway: $ovnGateway
    excludeIps: $ovnExcludeIps

EOF

kubectl apply -f network-prepare.yaml
sleep 2

ovnNet=$(kubectl get network | sed -n 2p | awk '{print $1}')
ovnProviderNet=$(kubectl get providernetwork | sed -n 2p | awk '{print $1}')
if [ -n "${ovnNet}" ] && [ -n "${ovnProviderNet}" ]
then
	echo "Networks created successfully"
else
	echo "Networks creation failed"
	exit 1
fi

echo "--------------------- Installing sdewan cnf with helm ---------------------"
curl https://helm.baltorepo.com/organization/signing.asc | sudo apt-key add -
sudo apt-get install apt-transport-https --yes
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm

envsubst < ./cnf/values.yaml >> ./cnf/value.yaml
mv ./cnf/value.yaml ./cnf/values.yaml
helm init
helm package ./cnf
helm install ./cnf-0.1.0.tgz
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

echo "--------------------- Setup Application ---------------------"

cat > httpbin-svc.yaml << EOF
---
apiVersion: v1
kind: Service
metadata:
  name: my-http-service
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $app_pod_name
spec:
  replicas: 1
  selector:
    matchLabels:
      name: simple-http-service
  template:
    metadata:
      labels:
        name: simple-http-service
        app: MyApp
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "ovn-networkobj"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [{"name": "ovn-network", "interface": "net0", "ipAddress": "$appLan0"}]}'
    spec:
      containers:
        - name: simple-http-service
          image: integratedcloudnative/httpbin:test
          ports:
            - containerPort: 80
          imagePullPolicy: IfNotPresent
          securityContext:
                  privileged: true
EOF
kubectl apply -f httpbin-svc.yaml
sleep 20

appStatus=$(kubectl get po | grep simple-http-service | awk '{print $3}')
if [ "$appStatus" == "Running" ]
then
	echo "Application $app_pod_name installation success"
else
        sleep 40
        appStatus=$(kubectl get po | grep simple-http-service | awk '{print $3}')
        if [ "$appStatus" != "Running" ]
        then
             echo "Application creation failed"
             exit 1
        fi
fi

echo "--------------------- Setup sdewan controller with helm ---------------------"
helm package ./controllers
helm install controllers-0.1.0.tgz
sleep 1m

echo "--------------------- Applying IPSec CRDs ---------------------"
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
kind: IpsecHost
metadata:
  name: ipsechost
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
    name: $edgeName
    remote: $hubIp
    pre_shared_key: test_key
    authentication_method: psk
    local_identifier: $cnfWan0
    crypto_proposal:
      - ipsecproposal
    force_crypto_proposal: "0"
    connections:
    - name: connA
      conn_type: tunnel
      mode: start
      local_sourceip: "%config"
      remote_subnet: 192.168.1.1/24,$hubIp/32
      crypto_proposal:
        - ipsecproposal

EOF

kubectl apply -f ipsec_config.yaml
sleep 10

export sdewan_cnf=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $1}' | head -1)
export sdwan_pod_virtual_ip=$(kubectl exec -it $sdewan_cnf ip address | grep $wan_interface | awk '/inet/{print $2}' | cut -f1 -d "/" | grep 192.168)

echo "--------------------- Adding firewall rules ---------------------"
cat > firewall_zone.yaml << EOF
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: FirewallZone
metadata:
  name: ovnnetwork
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
  network:
    - ovn-network
  input: ACCEPT
  output: ACCEPT
  forward: ACCEPT
EOF

kubectl apply -f firewall_zone.yaml

cat > firewall_zone_1.yaml << EOF
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: FirewallZone
metadata:
  name: pnetwork
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
  network:
    - pnetwork
  input: ACCEPT
  output: ACCEPT
  forward: REJECT
  masq: "0"
  mtu_fix: "1"
EOF

kubectl apply -f firewall_zone_1.yaml

echo "--------------------- Adding NAT rules ---------------------"
test -f $sdwanConfigs && . $sdwanConfigs
kubectl apply -f $sdwanConfigs.yaml

app_pod=$(kubectl get po | grep $app_pod_name | cut -f1 -d " ")
kubectl exec -it $app_pod ip route add 192.168.1.0/24 via $cnfLan0 dev net0

echo "Configuration finished"
