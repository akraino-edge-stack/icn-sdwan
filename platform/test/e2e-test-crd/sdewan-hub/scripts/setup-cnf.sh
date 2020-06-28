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

set -ex

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
kubectl delete -f sdewan-cnf.yaml
kubectl delete -f network-prepare.yaml
kubectl delete -f sdewan-controller.yaml
kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
[-f ipsec_config.yaml ] && kubectl delete -f ipsec_config.yaml
[-f ipsec_proposal.yaml ] && kubectl delete -f ipsec_proposal.yaml
}

error_detect()
{
	echo "Error on line $1"
	clean
}

trap "error_detect $LINENO" ERR

echo "--------------------- Setup CNF for sdewan hub -----------------------"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml --validate=false

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

kubectl create -f network-prepare.yaml
sleep 2

ovnProviderNet=$(kubectl get providernetwork | sed -n 2p | awk '{print $1}')
if [ -n "${ovnProviderNet}" ]
then
	echo "Network created successfully"
else
	echo "Network creation failed"
	exit 1
fi


echo "--------------------- Creating sdwan-cnf ---------------------"
cat > sdewan-cnf.yaml << EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $sdewan_cnf_name
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name 
  annotations:
spec:
  progressDeadlineSeconds: 600
  replicas: 2
  selector:
    matchLabels:
      sdewanPurpose: $sdewan_cnf_name 
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "ovn-networkobj"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [{"name": "pnetwork", "interface": "net1", "ipAddress": "$cnfWan0", "gateway": "$cnfWanGateway"}]}'
      labels:
        sdewanPurpose: $sdewan_cnf_name 
    spec:
      containers:
      - command:
              #- sleep
              #- "3600"
        - /bin/sh
        - /tmp/sdewan/entrypoint.sh
        image: integratedcloudnative/openwrt:dev
        imagePullPolicy: IfNotPresent
        name: sdewan
        readinessProbe:
          failureThreshold: 5
          httpGet:
            path: /
            port: 80
            scheme: HTTP
          initialDelaySeconds: 5
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 1
        securityContext:
          privileged: true
          procMount: Default
        volumeMounts:
        - mountPath: /tmp/sdewan
          name: sdewan-sh
          readOnly: true
        - mountPath: /tmp/podinfo
          name: podinfo
          readOnly: true
      nodeSelector:
        node-role.kubernetes.io/master: ""
      restartPolicy: Always
      volumes:
      - configMap:
          defaultMode: 420
          name: sdewan-sh
        name: sdewan-sh
      - name: podinfo
        downwardAPI:
          items:
            - path: "annotations"
              fieldRef:
                fieldPath: metadata.annotations
EOF

kubectl apply -f configmap.yaml
kubectl create -f sdewan-cnf.yaml
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

for name in $(kubectl get po | grep sdewan_cnf_name | awk '{print $1}')
do
	kubectl exec -it  $name /etc/init.d/network restart
done

echo "--------------------- Setup sdewan controller ---------------------"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml --validate=false
sleep 3m
kubectl apply -f sdewan-controller.yaml
sleep 30

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
  dh_group: modp4096
  encryption_algorithm: aes
  hash_algorithm: sha1

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
    name: sdewan-hub 
    remote: "%any" 
    pre_shared_key: test_key
    authentication_method: psk
    remote_identifier: host
    local_identifier: Hub
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

for name in $(kubectl get po | grep sdewan_cnf_name | awk '{print $1}')
do
        kubectl exec -it  $name /etc/init.d/ipsec restart
done

echo "--------------------- Configuration finished ---------------------"
