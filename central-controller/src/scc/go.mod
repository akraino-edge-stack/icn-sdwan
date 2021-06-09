module github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc

require (
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/go-playground/validator/v10 v10.4.1
	github.com/gorilla/mux v1.7.2
	github.com/jetstack/cert-manager v1.2.0
	github.com/matryer/runner v0.0.0-20190427160343-b472a46105b1
	github.com/open-ness/EMCO/src/orchestrator v0.0.0-00010101000000-000000000000
	github.com/open-ness/EMCO/src/rsync v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	go.etcd.io/etcd v3.3.12+incompatible
	google.golang.org/grpc v1.28.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/open-ness/EMCO/src/monitor => ../monitor
	github.com/open-ness/EMCO/src/orchestrator => ../vendor/github.com/open-ness/EMCO/src/orchestrator
	github.com/open-ness/EMCO/src/rsync => ../rsync
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
	k8s.io/kubectl => k8s.io/kubectl v0.19.0
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.1
)

go 1.14
