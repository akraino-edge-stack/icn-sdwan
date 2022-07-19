#!/bin/bash

#SPDX-License-Identifier: Apache-2.0
#Copyright (c) 2022 Intel Corporation

# usage: bash monitor-deploy.sh

set -ex

test -f ./monitor_configs && . monitor_configs
username=$(echo -n ${username-""} | base64 -)
token=$(echo -n ${token-""} | base64 -)
repo_name=$(echo -n ${repo_name-""} | base64 -)
cluster=$(echo -n ${cluster-""} | base64 -)
http_proxy=${http_proxy-""}
https_proxy=${https_proxy-""}

cat > monitor-secret.yaml << EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: monitor-git-monitor
  namespace: default
type: Opaque
data:
  username: $username
  token: $token
  repo: $repo_name
  clustername: $cluster
EOF

if [ -z "${username}" ]; then
	sed -i "s,{http_proxy},$http_proxy,g" monitor-deploy-pushmode.yaml
        sed -i "s,{https_proxy},$https_proxy,g" monitor-deploy-pushmode.yaml
	rm monitor-secret.yaml
	kubectl apply -f monitor-deploy-pushmode.yaml
else
	kubectl apply -f monitor-secret.yaml
	sed -i "s,{http_proxy},$http_proxy,g" monitor-deploy-pullmode.yaml
	sed -i "s,{https_proxy},$https_proxy,g" monitor-deploy-pullmode.yaml
	kubectl apply -f monitor-deploy-pullmode.yaml
fi



