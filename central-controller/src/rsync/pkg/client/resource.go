// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package client

import (
	"context"
	"errors"
	"fmt"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

var (
	// ErrLoggerInternal is returned when an error occurs due to the logger
	ErrLoggerInternal = errors.New("internal logger error")
)

// GetNodeLabels .. Fetch labels published by the K8s node
func (c *Client) GetNodeLabels(ctx context.Context) (map[string](map[string]string), error) {
	var nodeLabelMap = make(map[string](map[string]string))

	// iterate through Node Labels
	nodes, _ := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	for _, node := range nodes.Items {
		nodeLabelMap[node.Name] = node.Labels
	}

	return nodeLabelMap, nil
}

// GetClusterNodes lists all the nodes deployed on the cluster
func (c *Client) GetClusterNodes(ctx context.Context) (*corev1.NodeList, error) {

	log.Info("GetClusterNodes .. start", nil)

	// access the node APIs
	nodeList, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error("Kube client: Error in listing nodes", log.Fields{"Error": err.Error()})
		return nil, fmt.Errorf("Kube client: Error in listing nodes - %s", ErrLoggerInternal)
	}
	if len(nodeList.Items) < 1 {
		log.Warn("Can't find any Nodes in the cluster", nil)
		return nil, fmt.Errorf("Kube client: Can't find any Nodes in the cluster - %s", ErrLoggerInternal)
	}

	log.Info("GetClusterNodes .. end", log.Fields{"kube_cluster_name": nodeList.Items[0].ClusterName, "kube_nodes_count": len(nodeList.Items), "kube_node_1_name": nodeList.Items[0].GetName()})
	return nodeList, nil
}

// GetPodList lists all the pods deployed on the node
func (c *Client) GetPodList(ctx context.Context, nodeName string) (*corev1.PodList, error) {

	log.Info("Get Pods list", nil)
	// TODO: Call with the correct namespace. Empty string passed for default namespace in Pods()
	podList, err := c.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		log.Error("Kube client: Error in listing pods", log.Fields{"Error": err.Error()})
		return nil, fmt.Errorf("Kube client: Error in listing pods - %s", ErrLoggerInternal)
	}

	return podList, nil
}

// GetPodsTotalRequestsAndLimits ... Get resource total requests/limits (including running instances)
func (c *Client) GetPodsTotalRequestsAndLimits(ctx context.Context, podList *corev1.PodList) (map[corev1.ResourceName]resource.Quantity, map[corev1.ResourceName]resource.Quantity, error) {
	reqs, limits := map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}

	log.Info("Get pod total requests and limits", nil)

	for _, pod := range podList.Items {
		podReqs, podLimits := resourcehelper.PodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return reqs, limits, nil
}

// GetAvailableNodeResources returns the Resource available in the cluster(after dedecting running pods resource requirements)
func (c *Client) GetAvailableNodeResources(ctx context.Context, resName string) (int64, int64, map[string]int64, error) {
	var availableAfterResourceReqs int64
	var availableAfterResourceLimits int64

	log.Info("GetAvailableNodeResources .. start", log.Fields{"resName": resName})

	// Accounting for available resource on all nodes in the cluster
	nodeList, nodeErr := c.GetClusterNodes(ctx)
	if nodeErr != nil {
		log.Error("Unable to fetch Cluster nodes list", log.Fields{"err": nodeErr})
		return availableAfterResourceReqs, availableAfterResourceLimits, nil, fmt.Errorf("Unable to fetch Cluster nodes list - %s", nodeErr)
	}

	var nodeResourceMap = make(map[string]int64)
	for _, node := range nodeList.Items {
		// Get nonterminated pods list for the node
		podList, podErr := c.GetPodList(ctx, node.GetName())
		if podErr != nil {
			log.Error("Unable to fetch list of pods", log.Fields{"err": podErr})
			return availableAfterResourceReqs, availableAfterResourceLimits, nil, fmt.Errorf("Unable to initialize kube client - %s", podErr)
		}
		// Get all committed resources for pods
		reqs, limits, _ := c.GetPodsTotalRequestsAndLimits(ctx, podList)
		resReqs, resLimits := reqs[corev1.ResourceName(resName)], limits[corev1.ResourceName(resName)]

		// availableAfterResourceReqs is Resource count available(after deducting active pod and system daemon Resource requests)
		// availableAfterResourceLimits is Resource count available(after deducting active pod and system daemon Resource limits)
		// nodeResourceMap is the map of node to Resource count available(after deducting active pod and system daemon Resource requests)
		if val, ok := node.Status.Allocatable[corev1.ResourceName(resName)]; ok {
			valOrig := val
			val.Sub(resReqs)
			availableAfterResourceReqs += val.MilliValue() / 1000
			nodeResourceMap[node.Name] = val.MilliValue() / 1000
			availableAfterResourceLimits += val.MilliValue() / 1000

			log.Info("GetAvailableNodeResources info => ", log.Fields{
				"res_name":                  resName,
				"node_name":                 node.Name,
				"cluster_name":              node.GetClusterName(),
				"resAllocatedRequests":      valOrig,
				"resAllocatedToPodsRequest": resReqs,
				"resAllocatedToPodsLimit":   resLimits,
				"resAllocatedFinalRequests": val,
				"nodeResourceMap":           nodeResourceMap,
			})
		}
	}

	log.Info("GetAvailableNodeResources .. end",
		log.Fields{"res_name": resName,
			"availableAfterResourceReqs":   availableAfterResourceReqs,
			"availableAfterResourceLimits": availableAfterResourceLimits,
			"nodeResourceMap":              nodeResourceMap})

	return availableAfterResourceReqs, availableAfterResourceLimits, nodeResourceMap, nil
}
