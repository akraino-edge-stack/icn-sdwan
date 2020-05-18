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


HOME=$(pwd)

# Install dependencies
echo "Installing dependencies..."
sudo ./setup.sh -p libvirt

# Bring up 3 vms for edge-a, edge-b and sdewan-hub
cd $HOME
git clone http://gerrit.onap.org/r/multicloud/k8s
echo "Bringing up virtual machines for three clusters..."
for name in edge-a edge-b sdewan-hub
do
        cd $HOME/${name}
	echo "Start up cluster for ${name}..."
        vagrant up && vagrant up installer
done
sleep 700

# Checking vm status...
for name in edge-a edge-b sdewan-hub
do
        cd $HOME/${name}
        TIME_SPENT=0
	status=$(vagrant status | grep ${name} | awk '{print $2}')
	while [ $status != "running" ] && [ $TIME_SPENT < 300 ]
	do
	        status=$(vagrant status | grep ${name} | awk '{print $2}')
		sleep 50
		TIME_SPENT=$TIME_SPENT+50
	done
	if [ $status != "running" ]
	then
		echo "Virtual machine ${name} provision failed"
		clean
		exit 1
	else
		echo "Virtual machine ${name} provision success"
	fi
done



# Setup ipsec tunnels and applications
echo "Setup configs for the e2e scenario..."
for name in sdewan-hub edge-a edge-b
do
        cd $HOME/${name}
        vagrant ssh ${name}
	sudo -i
        cp -r /home/vagrant/scripts .
        cd scripts && ./setup-cnf.sh
done

echo "Testing the connectivity between applications..."
cd $HOME/edge-a
vagrant ssh edge-a
sudo -i
./scripts/test-connection.sh
sleep 3
clean

clean()
{
# Cleaning the env
echo "Cleaning the environment..."
echo "Deleting the vms..."
for name in edge-a edge-b sdewan-hub
do 
	cd $HOME/${name}
	vagrant destroy -f 
done
sleep 5
echo "Cleaning completed"
}


