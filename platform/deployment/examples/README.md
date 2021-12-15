# Example to verify
This is an example which you can test your SDEWAN after finishing deploying it.

## Pre-condition
**1.Install a simple nginx deployment and service**

`kubectl apply -f nginx-dp-svc.yaml`

**2.Apply the cnf service CR**

`kubectl apply -f cnfservice.yaml`

**3.Verify**

```
# From host, you can get the nginx response from cnf
curl <cnf_ip>:8866

# login to the cnf pod and see the iptables
kubectl exec -ti <cnf-pod-name> -n <namespace> -- sudo bash
iptable -L -t nat
# DNAT       tcp  --  anywhere             anywhere             tcp dpt:8866 to:<nginx-svc-ip>:80
```
