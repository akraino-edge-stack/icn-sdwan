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

test -f /home/vagrant/scripts/variables && . /home/vagrant/scripts/variables

sdewan_cnf=$(kubectl get po | grep $sdewan_cnf_name | awk '{print $1}' | head -1)
sdewan_virtual_ip=$(kubectl exec -it $sdewan_cnf ip address | grep $wan_interface | awk '/inet/{print $2}' | cut -f1 -d "/" | grep 192.168)
app_pod=$(kubectl get po | grep $app_pod_name | cut -f1 -d " ")
echo "Logging into the dummy pod $app_pod in edgeA..."


echo "Determine the ip address of remote host"
if [ "$sdewan_virtual_ip" == "192.168.1.5" ]
then
        remote_ip="192.168.1.6"
else
        remote_ip="192.168.1.5"
fi
echo "The remote ip is ${remote_ip}"

echo "Sending request to the remote httpbin. If the connection is established, it shall return the ip of the caller."
kubectl exec -it $app_pod -- curl -X GET "http://$remote_ip/ip" -H "accept: application/json" >> response.json
cat response.json

echo "Confirming the testing result..."
sudo apt install -y jq
rs=$(jq -r '.origin' response.json)
if [ "$rs" == "$sdewan_virtual_ip" ]
then
        echo "Ip matched. End-to-end test passed"
else
        echo "End-to-end test failed. Please check the logs for more details"
fi
