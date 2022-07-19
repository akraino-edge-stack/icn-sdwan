# Deploying monitor in clusters

Now monitor can be deployed in either push mode or pull mode.
If the monitor is deployed in pull mode, the user need to fill in the monitor_configs file with GitOps information including:
- username: the username used for the github/gitlab account
- token: the personal access token used for github/gitlab. User can apply thru https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
> Note: Token expiration time should be set less than 30 days. And scopes including repo, workflow, gist, project must be enabled for the access token
- repo: the repo name that the gitops configs will be uploaded to. e.g. "test-repo"
- clustername: the path where the gitops configs will be uploaded to. Please follow the 
format like this: akraino_scc_<overlay_name>+<device_name>

In the meanwhile, you can use the following command to install flux on your device(on the host)
Please execute both commands on your kubernetes master node.

## Prerequisites(ONLY needed in Pull mode)
### 1. Install flux components
```
brew install fluxcd/tap/flux
```

Note: If you need proxy to access github, you may need to customize flux like this:
```   
	apiVersion: kustomize.config.k8s.io/v1beta1
	kind: Kustomization
	resources:
	  - gotk-components.yaml
	  - gotk-sync.yaml
	patches:
	  - patch: |
	      apiVersion: apps/v1
	      kind: Deployment
	      metadata:
	        name: all
	      spec:
	        template:
	          spec:
	            containers:
	              - name: manager
	                env:
	                  - name: "HTTPS_PROXY"
		            value: "http://proxy.example.com:3129"
	                  - name: "NO_PROXY"
	                    value: ".cluster.local.,.cluster.local,.svc"      
	    target:
	      kind: Deployment
	      labelSelector: app.kubernetes.io/part-of=flux
```

### 2. Bootstrap flux repo
```
flux bootstrap github --owner=<username> --repository=<repo> --branch=<branch> --path=./clusters/<clustername> --personal
```

### 3. Fill in the monitor_configs file for GitOps information


## Install Monitor
**Use the monitor-deploy.sh to install the monitor components.**

```
bash monitor-deploy.sh
```
