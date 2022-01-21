# Helm Chart for cnf and controller

The installation has been verified on kubernetes v1.20.

## Pre-condition
**1.Install cert-manager**

`kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml`

**2.Label the node**

```
nodename=$(kubectl get node -o jsonpath='{.items[0].metadata.name}')
kubectl taint node $nodename node-role.kubernetes.io/master:NoSchedule-
kubectl label --overwrite node $nodename ovn4nfv-k8s-plugin=ovn-control-plane
```

**3.Install network**

For the network configuration, the helm charts of CNF and Controller need integrate Multus CNI with Calico as default network and icn-nodus. So you can refer to the [guide](https://github.com/akraino-edge-stack/icn-nodus/blob/master/doc/how-to-use.md#testing-with-cni-proxy) to setup your environment.

**4.Apply provide network**

- Create ovn-network and provider-network, e.g.
```
---
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: ProviderNetwork
metadata:
  name: pnetwork
spec:
  cniType: ovn4nfv
  ipv4Subnets:
  - subnet: 10.10.20.1/24
    name: subnet
    gateway: 10.10.20.1/24
    excludeIps: 10.10.20.2..10.10.20.9
  providerNetType: VLAN
  vlan:
    logicalInterfaceName: eno1.100 // Change to your interface name
    providerInterfaceName: eno1
    vlanId: "100"
    vlanNodeSelector: all

---
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: ovn-network
spec:
  # Add fields here
  cniType: ovn4nfv
  ipv4Subnets:
  - subnet: 172.16.30.1/24
    name: subnet1
    gateway: 172.16.30.1/24
```
- Update `helm/sdewan_cnf/values.yaml` to configure the network information

**5.Install helm**

```
curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
sudo apt-get install apt-transport-https --yes
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

## Steps to install CNF and CRD Controller

Please locate your directory to `./helm`.

**1.Create namespace for SDEWAN Central Controller v1Microservices**

`kubectl create namespace sdewan-system`

**2.Generate certificate for cnf**

`kubectl apply -f cert/cnf_cert.yaml`

**3.Install CNF**

```
helm package sdewan_cnf
helm install ./cnf-0.1.0.tgz --generate-name
```

**4.Install CRD controller**

```
helm package sdewan_controllers
helm install ./controllers-0.1.0.tgz --generate-name
```

