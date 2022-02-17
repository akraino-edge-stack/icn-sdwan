# Steps to register local cluster

**1. Copy Kubeconfig as admin.conf in this folder**
**2. Compiled as:**
   `$ go build -o reg_cluster ./reg_cluster.go`
**3. Edit config.json with mongo db and etcd ip**
**4. Run command to register cluster**
   `$ ./reg_cluster`

