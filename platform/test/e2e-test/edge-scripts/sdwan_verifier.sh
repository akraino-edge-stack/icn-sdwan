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


function login {
    login_url=http://$1/cgi-bin/luci/
    echo $(wget -S --spider --post-data "luci_username=root&luci_password=" $login_url 2>&1 | grep sysauth= | sed -r 's/.*sysauth=([^;]+);.*/\1/')
}

function call_get {
    url=$1
    token=$2
    echo $(curl -s -X "GET" -H "Cookie:sysauth=$token" -H "Content-Type: application/json" $url)
}

function call_post {
    url=$1
    token=$2
    payload=$3
    echo $(curl -s -X "POST" -H "Cookie:sysauth=$token" -H "Content-Type: application/json" -d "$payload" $url)
}

function call_put {
    url=$1
    token=$2
    payload=$3
    echo $(curl -s -X "PUT" -H "Cookie:sysauth=$token" -H "Content-Type: application/json" -d "$payload" $url)
}

function call_delete {
    url=$1
    token=$2
    echo $(curl -s -X "DELETE" -H "Cookie:sysauth=$token" -H "Content-Type: application/json" $url)
}

function wait_for_pod {
    status_phase=""
    while [[ "$status_phase" != "Running" ]]; do
        new_phase="$(kubectl get pods -o wide | grep ^$1 | awk '{print $3}')"
        if [[ "$new_phase" != "$status_phase" ]]; then
            status_phase="$new_phase"
        fi
        if [[ "$new_phase" == "Err"* ]]; then
            exit 1
        fi
    done
}

sdwan_pod_ip=$(kubectl get pods -o wide | grep $sdewan_cnf_name | awk '{print $6}')

echo "SDWAN pod ip:"$sdwan_pod_ip

echo "Login to sdwan ..."
security_token=""
while [[ "$security_token" == "" ]]; do
    echo "Get Security Token ..."
    security_token=$(login $sdwan_pod_ip)
    sleep 2
done    
echo "* Security Token: "$security_token

zone0="
{
	\"name\":\"lan0\",
	\"network\": [\"lan0\"],
	\"input\":\"ACCEPT\",
	\"output\":\"ACCEPT\",
	\"forward\":\"ACCEPT\"
}
"

zone1="
{
	\"name\":\"wan0\",
	\"network\": [\"wan0\"],
	\"input\":\"ACCEPT\",
	\"output\":\"ACCEPT\",
	\"forward\":\"REJECT\",
	\"masq\":\"0\",
	\"mtu_fix\":\"1\"
}
"

echo $(call_post http://$sdwan_pod_ip/cgi-bin/luci/sdewan/firewall/v1/zones $security_token "$zone0")
sleep 3
echo $(call_post http://$sdwan_pod_ip/cgi-bin/luci/sdewan/firewall/v1/zones $security_token "$zone1")
sleep 3
echo $(call_get http://$sdwan_pod_ip/cgi-bin/luci/sdewan/firewall/v1/zones $security_token)


echo "Set ipsec rules ..."
proposal="
{
    \"name\":\"my_test_proposal1\",
    \"encryption_algorithm\":\"aes128\",
    \"hash_algorithm\":\"sha256\",
    \"dh_group\":\"modp3072\"
}
"
echo $(call_post http://$sdwan_pod_ip/cgi-bin/luci/sdewan/ipsec/v1/proposals $security_token "$proposal")
sleep 3
echo "Proposals inserted..."
echo $(call_get http://$sdwan_pod_ip/cgi-bin/luci/sdewan/ipsec/v1/proposals $security_token)

rule="
{
    \"name\":\"$edgeName\",
    \"gateway\":\"$hubIp\",
    \"pre_shared_key\":\"test_key\",
    \"authentication_method\":\"psk\",
    \"local_identifier\":\"$cnfWan0\",
    \"crypto_proposal\":[\"my_test_proposal1\"],
    \"force_crypto_proposal\":\"0\",
    \"connections\":[{
        \"name\":\"connectionA\",
        \"type\":\"tunnel\",
        \"mode\":\"start\",
        \"local_sourceip\":\"%config\",
        \"remote_subnet\":\"192.168.1.1/24,$hubIp/32\",
        \"crypto_proposal\":[\"my_test_proposal1\"]
    }]
}
"
echo $(call_post http://$sdwan_pod_ip/cgi-bin/luci/sdewan/ipsec/v1/sites $security_token "$rule")
sleep 3
echo "IPSec configs inserted..."
echo $(call_get http://$sdwan_pod_ip/cgi-bin/luci/sdewan/ipsec/v1/sites $security_token)

echo "Restarting IPSec service..."
operation="
{
	\"action\":\"restart\"
}
"
echo $(call_put http://$sdwan_pod_ip/cgi-bin/luci/sdewan/v1/service/ipsec $security_token "$operation")

export sdwan_pod_virtual_ip=$(kubectl exec -it $sdewan_cnf_name ip address | grep $wan_interface | awk '/inet/{print $2}' | cut -f1 -d "/" | grep 192.168)

echo "Applying sdwan redirect configs"
test -f ./$sdwanConfigs && . $sdwanConfigs
redirect=${redirect}

echo $(call_post http://$sdwan_pod_ip/cgi-bin/luci/sdewan/firewall/v1/redirects $security_token "$redirect")
sleep 3
echo $(call_get http://$sdwan_pod_ip/cgi-bin/luci/sdewan/firewall/v1/redirects $security_token)

echo "Restarting Firewall service..."
operation="
{
        \"action\":\"restart\"
}
"
echo $(call_put http://$sdwan_pod_ip/cgi-bin/luci/sdewan/v1/service/firewall $security_token "$operation")


echo "Add route to bring the app traffic to sdewan cnf"
export app_pod=$(kubectl get po | grep $app_pod_name | cut -f1 -d " ")
kubectl exec -it $app_pod ip route add $internal_subnet via $cnfLan0 dev net0

echo "Test Completed!"
