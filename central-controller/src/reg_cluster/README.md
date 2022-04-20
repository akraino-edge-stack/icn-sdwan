# Steps to register local cluster

**1. Copy Kubeconfig as admin.conf in this folder**
  You can name it by default as 'admin.conf' or define it using --kubeconfigPath
  while executing the command.
**2. Compiled as:**
   `$ go build -o reg_cluster ./reg_cluster.go`
**3. Edit config.json with mongo db and etcd ip**
**4. Run command to register cluster**
   Note: You shall have the secrets used for mongo already created
   in the sdewan-system namespace
   `$ ./reg_cluster` This will use the default secret names
   Or
   `$ ./reg_cluster --mongoSecret <mongo-secret-name> --mongoDataSecret
    <mongo-data-secret-name>`

