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


BASE=$(pwd)


clean()
{
# Cleaning the env
echo "Cleaning the environment..."
echo "Deleting the vms..."
for name in edge-a edge-b sdewan-hub
do
	cd $BASE/${name}
	vagrant destroy -f &
        sleep 10
done
echo "Cleaning completed"
}

error_report()
{
	echo "Error on line $1"
	echo "End-to-end test failed"
}

trap 'error_report $LINENO' ERR
#trap 'clean' EXIT

base()
{
# Install dependencies
echo "Installing dependencies..."
sudo ./setup.sh -p libvirt

# Bring up 3 vms for edge-a, edge-b and sdewan-hub
cd $BASE
git clone http://gerrit.onap.org/r/multicloud/k8s && cd k8s
echo "Bringing up virtual machines for three clusters..."
for name in edge-a edge-b sdewan-hub
do
        cd $BASE/${name}
	echo "Start up cluster for ${name}..."
        vagrant up && vagrant up installer
        sleep 40
done


# Checking vm status...
for name in edge-a edge-b sdewan-hub
do
        cd $BASE/${name}
	vagrant ssh ${name} -- -t 'mkdir -p /home/vagrant/.kube; sudo cp -i /etc/kubernetes/admin.conf /home/vagrant/.kube/config; sudo chown $(id -u):$(id -g) $HOME/.kube/config'
	Status=$(vagrant ssh ${name} -- -t 'kubectl get po -n operator  | grep 'nfn-agent'' | grep 'nfn-agent' | awk '{print $3}')
	if [ $Status != "Running" ]
	then
		echo "Virtual machine ${name} provision failed"
		exit 1
	else
		echo "Virtual machine ${name} provision success"
	fi
done
}

# Setup ipsec tunnels and applications
echo "Setup configs for the e2e scenario..."
for name in sdewan-hub edge-a edge-b
do
        cd $BASE/${name}
        vagrant ssh ${name} -- -t 'cd /home/vagrant/scripts; ./setup-cnf.sh'
done


echo "Testing the connectivity between applications..."
cd $BASE/edge-a
vagrant ssh edge-a -- -t './scripts/test-connection.sh'
sleep 3
