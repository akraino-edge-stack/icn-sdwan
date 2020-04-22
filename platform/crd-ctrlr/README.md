# Sdewan operator

The sdewan operator is developed under kubebuilder framework

## Deployment Guide

The API admission webhook depends on cert-manager so we need to install cert-manager first.

We have the image built and published at `integratedcloudnative/sdewan-controller:dev`. The openwrt
docker image we used for test is at `integratedcloudnative/openwrt:dev`. To use some other images,
we need to make changes in deployment yaml file.

The installation steps for Sdewan operator:
1. kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
2. kubectl apply -f platform/crd-ctrlr/examples/sdewan-controller.yaml

Sample deployment of CNF:
1. Setup ovn networks
  ```
  kubectl apply -f examples/attach-network-ovn.yaml  examples/ovn-net1.yaml  examples/ovn-net2.yml
  ```
2. Launch CNF deployment
  ```
  kubectl apply -f examples/cnf-deployment.yaml
  ```
3. Create rule for the CNF
  ```
  kubectl apply -f src/config/samples/batch_v1alpha1_mwan3policy.yaml
  ```


## Developer Guide

Project initialization(mostly, developers should not execute this step)
```
go mod init sdewan.akraino.org/sdewan
kubebuilder init --domain sdewan.akraino.org
```

To create new CRD and controller
```
kubebuilder create api --group batch --version  v1alpha1  --kind  Mwan3Policy
```

To run local controller without webhook(For test/debug purpose)
```
make install
make run ENABLE_WEBHOOKS=false
```

To build controller docker image
```
make docker-build IMG="integratedcloudnative/sdewan-controller:dev"
```

To generate yaml file for controller deployment
```
make gen-yaml IMG="integratedcloudnative/sdewan-controller:dev"
```

### Controller Implementation

![sdewan_dev](diagrams/sdewan_dev.png)

- One CRD one controller
- Controller watches itself CR and the Deployment(ready status only)
- Reconcile calls WrtProvider to add/update/delete rules for CNF
- CnfProvider interfaces defines the function CNF function calls. WrtProvider is one implementation of CnfProvider
- For the users, CNF rules are CRs. But for openwrt, the rules are openwrt rule entities. We can pass the CRs to OpenWRT API. Instead, we need to convert the CRs to OpenWRT entities.
- Finalizer should be added to CR only when AddUpdate call succeed. Likewise, finalizer should be removed from CR only when Delete call succeed.
- **As we have many CRDs, so there could be many duplicate code. For example, convertCrd, AddUpdateXX, and even reconcile logic. So we need to extract the similar logic into functions to reduce the duplicatioin.**

### What we have implemented

- CNF image built from HuiFeng's script. I have uploaded the image at `integratedcloudnative/openwrt:dev`
- The CNF sample deployment yaml file under sample directory (together with configmap and ovn network yaml files)
- A runable framework with Mwan3Policy CRD and controller implemented. It means we can run the controller and add/update/delete mwan3policy rules.

### What we don't have yet

- Add a watch for deployment, so that the controller can get the CNF ready status change. [predicate feature](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/predicate#example-Funcs) should be used to filter no-status event.
- Implemente the remain CRDs/controllers. As all the controller logics are almost the same, some workload will be the extracting of the similar logic and make them functions.
- Add raw webhook to implemente the label based permission system
- Add validation webhook to validate CR

## References

- https://book.kubebuilder.io/
- https://openwrt.org/
