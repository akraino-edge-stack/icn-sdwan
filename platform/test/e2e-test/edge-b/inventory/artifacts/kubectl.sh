#!/bin/bash
kubectl --kubeconfig=${BASH_SOURCE%/*}/admin.conf $@
