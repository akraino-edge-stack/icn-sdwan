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
kubectl delete -f sdewan-cnf.yaml
kubectl delete -f network-prepare.yaml
}

error_detect()
{
	echo "Error on line $1"
	clean
}

trap "error_detect $LINENO" ERR

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

kubectl create -f network-prepare.yaml
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
    config interface 'lan0'
        option ifname 'net0'
        option proto 'static'
        option ipaddr '$cnfLan0'
        option netmask '255.255.255.0'

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
    k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [{"name": "ovn-network", "interface": "net0", "ipAddress": "$cnfLan0"}, {"name": "pnetwork", "interface": "net1", "ipAddress": "$cnfWan0", "gateway": "$cnfWanGateway"}]}'
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
sleep 20

sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}')
if [ "$sdwan_status" == "Running" ]
then
	echo "Sdewan cnf $sdewan_cnf_name created successfully"
else
        sleep 40
	sdwan_status=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $3}')
	if [ "$sdwan_status" != "Running" ]
	then
	     echo "Sdewan cnf creation failed"
             exit 2
        fi
fi

kubectl exec -it  $sdewan_cnf_name /etc/init.d/network restart


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
kubectl create -f httpbin-svc.yaml
sleep 20

appStatus=$(kubectl get po | grep simple-http-service | awk '{print $3}')
if [ "$appStatus" == "Running" ] 
then
	echo "Application $app_pod_name installation success"
else
        sleep 20
        appStatus=$(kubectl get po | grep simple-http-service | awk '{print $3}')
        if [ "$appStatus" != "Running" ]
        then
             echo "Application creation failed"
             exit 1
        fi
fi

echo "Setup openwrt configurations"
bash sdwan_verifier.sh
