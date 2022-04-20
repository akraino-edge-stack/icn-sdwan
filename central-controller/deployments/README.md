# Steps for running v1 API microservices

### Precondition
**1. Install cert-manager**

`$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml`

### Steps to install packages from yaml
**1. Create namespace for SDEWAN Central Controller v1Microservices**

`$cd kubernetes`
`$ kubectl create namespace sdewan-system`

**2. Create Databases used by SDEWAN Central Controller v1 Microservices for Etcd and Mongo**

`$ kubectl apply -f scc_secret.yaml -n sdewan-system`
`$ kubectl apply -f scc_etcd.yaml -n sdewan-system`
`$ kubectl apply -f scc_mongo.yaml -n sdewan-system`

**3. create SDEWAN Central Controller v1 Microservices**

`$ kubectl apply -f scc.yaml -n sdewan-system`

`$ kubectl apply -f scc_rsync.yaml -n sdewan-system`

**4. install monitor resources**

`$ kubectl apply -f monitor-deploy.yaml -n sdewan-sysyem`
