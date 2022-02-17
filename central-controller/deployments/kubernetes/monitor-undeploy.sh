kubectl delete -f monitor/crds/k8splugin_v1alpha1_resourcebundlestate_crd.yaml
kubectl delete -f monitor/role.yaml
kubectl delete -f monitor/cluster_role.yaml
kubectl delete -f monitor/role_binding.yaml
kubectl delete -f monitor/clusterrole_binding.yaml
kubectl delete -f monitor/service_account.yaml
kubectl delete -f monitor/operator.yaml
