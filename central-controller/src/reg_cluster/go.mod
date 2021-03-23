go 1.14

module central-controller/src/reg_cluster

require (
        github.com/google/uuid v1.1.2 // indirect
        github.com/open-ness/EMCO/src/monitor v0.0.0-00010101000000-000000000000
        github.com/open-ness/EMCO/src/orchestrator v0.0.0-00010101000000-000000000000
        github.com/open-ness/EMCO/src/rsync v0.0.0-00010101000000-000000000000
        go.etcd.io/etcd v3.3.12+incompatible
        google.golang.org/grpc v1.28.0
        k8s.io/client-go v12.0.0+incompatible
)

replace (
        github.com/open-ness/EMCO/src/monitor => ../monitor
        github.com/open-ness/EMCO/src/rsync => ../rsync
        github.com/open-ness/EMCO/src/orchestrator => ../vendor/github.com/open-ness/EMCO/src/orchestrator
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
