```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```

# SDEWAN crd-ctrlr examples 

## To deploy an example CNF

```shell
kubectl apply -f attach-network-ovn.yaml
kubectl apply -f ovn-net1.yaml ovn-net2.yml ovn-provnet.yaml sdewan-cm.yaml
kubectl apply -f sdewan-deployment.yaml
```
