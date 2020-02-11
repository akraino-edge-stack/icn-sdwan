# Sdewan operator

The sdewan operator is developed under operator-framework.

Deploying the operator creates a pod running the Sdewan controller,
which watches the Sdewan instances. For each created Sdewan instance,
the controller creates a pod, a configmap and a service for the instance.

The pod runs openswrt which provides network services, i.e. sdwan, firewall
, ipsec etc. We begins with supporting sdwan service.

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

In this patch, we also do the following things for mwan3 service:
1. Add the Mwan3Rule CRD. For now, it does have much content besides the policy
information.
