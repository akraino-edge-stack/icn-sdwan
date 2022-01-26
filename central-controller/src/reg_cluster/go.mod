go 1.17

module central-controller/src/reg_cluster

require (
	github.com/open-ness/EMCO/src/orchestrator v0.0.0-00010101000000-000000000000
	github.com/open-ness/EMCO/src/rsync v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
)

require (
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.3 // indirect
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
