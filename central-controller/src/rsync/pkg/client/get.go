// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ReadResource is structure used for reading resource
type ReadResource struct {
	Gvk       schema.GroupVersionKind `json:"GVK,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Namespace string                  `json:"namespace,omitempty"`
}

// Get gets the resource from the remote cluster
func (c *Client) Get(gvkRes []byte, namespace string) ([]byte, error) {

	var g ReadResource
	var ns string

	err := json.Unmarshal(gvkRes, &g)
	if err != nil {
		return nil, fmt.Errorf("Invalid read resource %v", err)
	}

	if namespace == "default" && g.Namespace != "" {
		ns = g.Namespace
	} else {
		ns = namespace
	}
	// Create a mapper for the GVK
	mapping, err := c.RestMapper.RESTMapping(schema.GroupKind{
		Group: g.Gvk.Group,
		Kind:  g.Gvk.Kind,
	}, g.Gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("RESTMapping for GVK failed %v", err)
	}
	gvr := mapping.Resource
	opts := metav1.GetOptions{}
	var unstruct *unstructured.Unstructured
	// Handle based on namespace scopped GVK or not
	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		unstruct, err = c.DynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), g.Name, opts)
	case meta.RESTScopeNameRoot:
		unstruct, err = c.DynamicClient.Resource(gvr).Get(context.TODO(), g.Name, opts)
	default:
		return nil, fmt.Errorf("RESTScopeName for GVK failed %v, %s", err, g.Gvk.String())
	}
	if err != nil {
		return nil, fmt.Errorf("Getting getting RESTScopeName %v", err)
	}

	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Failed unstruct MarshalJSON %v", err)
	}
	return b, nil
}
