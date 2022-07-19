// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"encoding/base64"
	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"strings"
)

var KubeConfigNotFoundErr = pkgerrors.New("Get kubeconfig failed")

// DecodeYAMLData reads a string to extract the Kubernetes object definition
func DecodeYAMLData(data string, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(data), nil, into)
	if err != nil {
		log.Error("Error DecodeYAMLData", log.Fields{"error": err})
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

// GetKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
var GetKubeConfig = func(clustername string, level string, namespace string) ([]byte, error) {
	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}

	ccc := db.NewCloudConfigClient()

	log.Info("Querying CloudConfig", log.Fields{"strs": strs, "level": level, "namespace": namespace})
	cconfig, err := ccc.GetCloudConfig(strs[0], strs[1], level, namespace)
	if err != nil {
		return nil, KubeConfigNotFoundErr
	}
	log.Info("Successfully looked up CloudConfig", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

	dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

var GetGitOpsConfig = func(clustername string, level string, namespace string) (mtypes.GitOpsSpec, error) {
	if !strings.Contains(clustername, "+") {
		return mtypes.GitOpsSpec{}, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return mtypes.GitOpsSpec{}, pkgerrors.New("Not a valid cluster name")
	}
	ccc := db.NewCloudConfigClient()

	cfg, err := ccc.GetGitOpsConfig(strs[0], strs[1], level, namespace)
	if err != nil {
		return mtypes.GitOpsSpec{}, err
	}
	return cfg.Config, nil
}
