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

sdewan_virtual_ip=$sdwan_pod_virtual_ip
echo "Logging into the dummy pod $app_pod in edgeA..."
kubectl exec -it $app_pod bash


echo "Determine the ip address of remote host"
if [ "$sdewan_virtual_ip" == "192.168.1.5" ]
then 
	remote_ip="192.168.1.6"
else
	remote_ip="192.168.1.5"
fi
echo "The remote ip is ${remote_ip}"

echo "Sending request to the remote httpbin..."
curl -X GET "http://$remote_ip/ip" -H "accept: application/json" -o response.json
cat response.json

echo "Confirming the testing result..."
apt install -y jq
rs=$(jq -r '.origin' response.json)
if [ "$rs" == "$sdwan_pod_virtual_ip" ]
then
	echo "End-to-end test passed"
else
	echo "End-to-end test failed. Please check the logs for more details"
