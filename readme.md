# Sdewan operator

The sdewan operator is developed under kubebuilder framework

We define two CRDs in this patch: Sdewan and Mwan3Conf.

Sdewan defines the CNF base info, which node we should deploy the CNF on,
which network should the CNF use with multus CNI, etc.

The Mwan3Conf defines the mwan3 rules. In the next step, we are going to
develop the firewall and the ipsec functions. Mwan3Conf is validated by k8s
api admission webhook.

For each created Sdewan instance, the controller creates a pod, a configmap
and a service for the instance. The pod runs openswrt which provides network
services, i.e. sdwan, firewall, ipsec etc.

The configmap stores the network interface information and the entrypoint.sh.
The network interface information has the following format:
```
[
  {
    "name": "ovn-priv-net",
    "isProvider": false,
    "interface": "net0",
    "defaultGateway": false
  }
]
```

The service created by the controller is used for openwrt api access.
We call this svc to apply rules, get openwrt info, restart openwrt service.

After the openwrt pod ready, the Sdewan controller apply the configured mwan3 rules.
mwan3 rule details are configured in Mwan3Conf CR, which is referenced by Sdewan.Spec.Mwan3Conf
Every time the Mwan3Conf instance changes, the controller re-apply the new rules by calling opwnrt
api. We can also change the rule refernce at the runtime.

## Deployment

The API admission webhook depends on cert-manager so we need to install cert-manager first.

To install the CRD and the controller, we can follow this guide.
https://book.kubebuilder.io/cronjob-tutorial/running-webhook.html

We have the image built and published at `integratedcloudnative/sdewan-controller:dev`. The openwrt
docker image we used for test is at `integratedcloudnative/openwrt:dev`. To use some other images,
we need to make configuration in `config/default/manager_image_patch.yaml`

The simple installation steps:
1. kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
2. kubectl apply -f sdewan-deploy.yaml

## References

- https://book.kubebuilder.io/
- https://openwrt.org/
