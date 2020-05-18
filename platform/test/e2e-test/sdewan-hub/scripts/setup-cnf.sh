#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################


set -ex
set -o errexit
set -o nounset
set -o pipefail


test -f ./variables 
. variables
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
}

error_detect()
{
	echo "Error on line $1"
	clean
}

trap 'error_detect $LINENO' ERR
trap 'clean' EXIT

echo "Creating ovn networks..."
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


echo "Creating sdwan-cnf ..."
cat > sdewan-cnf.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: sdwan-config
data:
  entrypoint.sh: |
    #!/bin/bash
    # Always exit on errors.
    set -e

    cat > /etc/config/network << EOF
    config interface 'wan0'
        option ifname 'net1'
        option proto 'static'
        option ipaddr '$cnfWan0'
        option netmask '255.255.255.0'
    EOF
    
    iptables -t nat -L
    /sbin/procd &
    /sbin/ubusd &
    sleep 1
    /etc/init.d/rpcd start
    /etc/init.d/dnsmasq start
    /etc/init.d/network start
    /etc/init.d/odhcpd start
    /etc/init.d/uhttpd start
    /etc/init.d/log start
    /etc/init.d/dropbear start
    /etc/init.d/mwan3 restart

    echo "Entering sleep... (success)"

    # Sleep forever.
    while true; do sleep 100; done

---
apiVersion: v1
kind: Pod
metadata:
  name: $sdewan_cnf_name
  annotations:
    k8s.v1.cni.cncf.io/networks: '[{ "name": "ovn-networkobj"}]'
    k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [{"name": "pnetwork", "interface": "net1", "ipAddress": "$cnfWan0", "gateway": "$cnfWanGateway"}]}'
spec:
  containers:
  - name: sdwan-pod-net
    image: integratedcloudnative/openwrt:test
    ports:
      - containerPort: 22
      - containerPort: 80
    command:
    - /bin/sh
    - /init/entrypoint.sh
    imagePullPolicy: IfNotPresent
    securityContext:
      privileged: true
    volumeMounts:
      - name: entrypoint-sh
        mountPath: /init
  volumes:
    - name: entrypoint-sh
      configMap:
        name: sdwan-config
        items:
        - key: entrypoint.sh
          path: entrypoint.sh
EOF

kubectl create -f sdewan-cnf.yaml
sleep 10

sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}')
if [ "$sdwan_status" == "Running" ]
then
	echo "Sdewan cnf $sdewan_cnf_name created successfully"
else
        sleep 50
	sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}')
	if [ "$sdwan_status" != "Running" ]
	then
		echo "Sdewan cnf $sdewan_cnf_name creation failed"
		exit 2
	fi
fi

kubectl exec -it  $sdewan_cnf_name /etc/init.d/network restart


echo "Setup openwrt configurations"
bash sdwan_verifier.sh


#clean

