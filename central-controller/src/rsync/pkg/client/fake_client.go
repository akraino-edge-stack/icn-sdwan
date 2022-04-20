// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package client

import (
	"context"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// NewKubeFakeClient returns a fake kube client object
func NewKubeFakeClient() (*Client, error) {
	log.Info("NewKubeFakeClient .. start", nil)

	var kl Client
	kl.Clientset = fake.NewSimpleClientset()

	// populate mock node1 with resources for cluster-1
	var node1 corev1.Node
	node1.Name = "emco-cluster-node-1"
	node1.ClusterName = "emco-cluster-1"
	resList := "cpu=10,qat.intel.com/generic=10,memory=5G,pods=10,services=7"
	nodeLabels := map[string]string{"feature.node.kubernetes.io/intel_qat": "true", "feature.node.kubernetes.io/pci-0300_1a03.present": "true", "feature.node.kubernetes.io/intel_qat/v3": "true"}
	resourceQuotaSpecList, _ := PopulateResourceListV1WithDefValues(resList)
	node1.Status.Allocatable = resourceQuotaSpecList
	node1.Status.Capacity = resourceQuotaSpecList
	node1.Labels = nodeLabels
	kl.Clientset.CoreV1().Nodes().Create(context.TODO(), &node1, metav1.CreateOptions{})

	// populate mock node2 with resources for cluster-1
	var node2 corev1.Node
	node2.Name = "emco-cluster-node-2"
	node2.ClusterName = "emco-cluster-1"
	resList = "cpu=5,qat.intel.com/generic=15,memory=10G,pods=5,services=5"
	nodeLabels = map[string]string{"feature.node.kubernetes.io/intel_qat": "true", "feature.node.kubernetes.io/pci-0300_1a03.present": "true", "feature.node.kubernetes.io/intel_qat.device": "c6xx"}
	resourceQuotaSpecList, _ = PopulateResourceListV1WithDefValues(resList)
	node2.Status.Allocatable = resourceQuotaSpecList
	node2.Status.Capacity = resourceQuotaSpecList
	node2.Labels = nodeLabels
	kl.Clientset.CoreV1().Nodes().Create(context.TODO(), &node2, metav1.CreateOptions{})

	// populate mock node1 with resources for cluster-2
	var node3 corev1.Node
	node3.Name = "emco-cluster-2-node-1"
	node3.ClusterName = "emco-cluster-2"
	resList = "cpu=15,qat.intel.com/generic=2,memory=1G,pods=5,services=2"
	nodeLabels = map[string]string{"feature.node.kubernetes.io/intel_qat": "true", "feature.node.kubernetes.io/pci-0300_1a03.present": "true", "feature.node.kubernetes.io/intel_qat/v3": "true"}
	resourceQuotaSpecList, _ = PopulateResourceListV1WithDefValues(resList)
	node3.Status.Allocatable = resourceQuotaSpecList
	node3.Status.Capacity = resourceQuotaSpecList
	node3.Labels = nodeLabels
	kl.Clientset.CoreV1().Nodes().Create(context.TODO(), &node3, metav1.CreateOptions{})

	// populate mock node2 with resources for cluster-2
	var node4 corev1.Node
	node4.Name = "emco-cluster-2-node-2"
	node4.ClusterName = "emco-cluster-2"
	resList = "cpu=10,qat.intel.com/generic=1,memory=2G,pods=2,services=2"
	nodeLabels = map[string]string{"feature.node.kubernetes.io/intel_qat": "true", "feature.node.kubernetes.io/pci-0300_1a03.present": "true", "feature.node.kubernetes.io/intel_qat.device": "c6xx"}
	resourceQuotaSpecList, _ = PopulateResourceListV1WithDefValues(resList)
	node4.Status.Allocatable = resourceQuotaSpecList
	node4.Status.Capacity = resourceQuotaSpecList
	node4.Labels = nodeLabels
	kl.Clientset.CoreV1().Nodes().Create(context.TODO(), &node4, metav1.CreateOptions{})

	log.Info("NewKubeFakeClient .. end", nil)
	return &kl, nil
}
