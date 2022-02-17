kubectl apply -f monitor/crds/k8splugin_v1alpha1_resourcebundlestate_crd.yaml
kubectl apply -f monitor/role.yaml
kubectl apply -f monitor/cluster_role.yaml
kubectl apply -f monitor/role_binding.yaml
kubectl apply -f monitor/clusterrole_binding.yaml
kubectl apply -f monitor/service_account.yaml
kubectl apply -f monitor/operator.yaml
