// SPDX-License-Identifier: Apache-2.0
// Based on Code: https://github.com/johandry/klient

package client

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a namespace with the given name
func (c *Client) CreateNamespace(namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"name": namespace,
			},
		},
	}
	_, err := c.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	// if errors.IsAlreadyExists(err) {
	// 	// If it failed because the NS is already there, then do not return such error
	// 	return nil
	// }

	return err
}

// DeleteNamespace deletes the namespace with the given name
func (c *Client) DeleteNamespace(namespace string) error {
	return c.Clientset.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}

// NodesReady returns the number of nodes ready
func (c *Client) NodesReady() (ready int, total int, err error) {
	nodes, err := c.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, 0, err
	}
	total = len(nodes.Items)
	if total == 0 {
		return 0, 0, nil
	}
	for _, n := range nodes.Items {
		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				ready++
				break
			}
		}
	}

	return ready, len(nodes.Items), nil
}

// GetMasterNodeIP returns the master node IP of the deployed app
func (c *Client) GetMasterNodeIP() (nodeIP string, err error) {

	ip := ""
	nodes, err := c.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ip, err
	}
	total := len(nodes.Items)
	if total == 0 {
		return ip, nil
	}
	for _, item := range nodes.Items {
		for _, address := range item.Status.Addresses {
			// This gives the ip address of the master node
			ip = address.Address
			break
		}
		break
	}

	return ip, nil
}

// Version returns the cluster version. It can be used to verify if the cluster
// is reachable. It will return an error if failed to connect.
func (c *Client) Version() (string, error) {
	cl, ok := (c.Clientset.(*kubernetes.Clientset))
	if ok {
		v, err := cl.ServerVersion()
		if err != nil {
			return "", err
		}
		return v.String(), nil
	}
	return "", errors.New("error in getting client")
}
