# Steps to setup Overlay Controller cluster

###Prerequisite
**Install Kubernetes**

###Install dependencies
**1. Install cert-manager**

`$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml`

**2. Create namespace for SDEWAN Overlay Controller Microservices**

`$ kubectl create namespace sdewan-system`

###Install CNF
**Please follow the README.md under ../platform/cnf folder for installation**

###Install CRD Controller
**Please follow the README.md under ../platform/crd-ctrlr folder for installation**

###Install Overlay Controller Microservices
**Please follow the README.md under deployments folder for installation**

###Configurations for Overlay Controller
**1. Routing rule**

**2. Local cluster registration**
**Please follow the README.md under reg_cluster/ to finish the registration**
