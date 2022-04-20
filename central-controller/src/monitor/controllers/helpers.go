// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceProvider interface {
	GetClient() client.Client
	UpdateStatus(*k8spluginv1alpha1.ResourceBundleState, runtime.Object) (bool, error)
	DeleteObj(*k8spluginv1alpha1.ResourceBundleState, string) bool
}

// checkLabel verifies if the expected label exists and returns bool
func checkLabel(labels map[string]string) bool {

	_, ok := labels["emco/deployment-id"]
	if !ok {
		return false
	}
	return true
}

// returnLabel verifies if the expected label exists and returns a map
func returnLabel(labels map[string]string) map[string]string {

	l, ok := labels["emco/deployment-id"]
	if !ok {
		return nil
	}
	return map[string]string{
		"emco/deployment-id": l,
	}
}

// listResources lists resources based on the selectors provided
// The data is returned in the pointer to the runtime.Object
// provided as argument.
func listResources(cli client.Client, namespace string,
	labelSelector map[string]string, returnData client.ObjectList) error {

	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(labelSelector),
	}

	err := cli.List(context.TODO(), returnData, listOptions)
	if err != nil {
		log.Printf("Failed to list CRs: %v", err)
		return err
	}

	return nil
}

// listClusterResources lists non-namespace resources based
// on the selectors provided.
// The data is returned in the pointer to the runtime.Object
// provided as argument.
func listClusterResources(cli client.Client,
	labelSelector map[string]string, returnData client.ObjectList) error {
	return listResources(cli, "", labelSelector, returnData)
}

func GetCRListForResource(client client.Client, item *unstructured.Unstructured) (*k8spluginv1alpha1.ResourceBundleStateList, error) {
	rbStatusList := &k8spluginv1alpha1.ResourceBundleStateList{}

	// Find the CRs which track this resource via the labelselector
	crSelector := returnLabel(item.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this resource")
		return rbStatusList, fmt.Errorf("Unexpected Error: Resource not filtered by predicate")
	}
	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	err := listResources(client, item.GetNamespace(), crSelector, rbStatusList)
	if err != nil {
		return rbStatusList, err
	}
	if len(rbStatusList.Items) == 0 {
		return rbStatusList, nil
	}
	return rbStatusList, nil
}

func UpdateCR(c client.Client, item *unstructured.Unstructured, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) error {
	var err error
	var found bool
	rbStatusList, err := GetCRListForResource(c, item)
	if err != nil {
		return err
	}

	for _, cr := range rbStatusList.Items {
		orgStatus := cr.Status.DeepCopy()

		// Not scheduled for deletion
		if item.GetDeletionTimestamp() == nil {
			_, err := UpdateStatus(&cr, item)
			if err != nil {
				fmt.Println("Error updating CR")
				return err
			}
			// Commit
			err = CommitCR(c, &cr, orgStatus)
		} else {
			// Scheduled for deletion
			found, err = DeleteObj(&cr, namespacedName.Name, gvk)
			if found && err == nil {
				err = CommitCR(c, &cr, orgStatus)
			}
		}
	}
	return err
}

func UpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, item *unstructured.Unstructured) (bool, error) {

	switch item.GetObjectKind().GroupVersionKind() {
	case schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}:
		return ConfigMapUpdateStatus(cr, item)
	case schema.GroupVersionKind{Version: "v1", Kind: "Service"}:
		return ServiceUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}:
		return DaemonSetUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}:
		return DeploymentUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}:
		return JobUpdateStatus(cr, item)
	case schema.GroupVersionKind{Version: "v1", Kind: "Pod"}:
		return PodUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}:
		return StatefulSetUpdateStatus(cr, item)
	case schema.GroupVersionKind{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}:
		return CsrUpdateStatus(cr, item)
	}
	return false, fmt.Errorf("Resource not supported explicitly")
}

func DeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string, gvk schema.GroupVersionKind) (bool, error) {

	switch gvk {
	case schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}:
		return ConfigMapDeleteObj(cr, name)
	case schema.GroupVersionKind{Version: "v1", Kind: "Service"}:
		return ServiceDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}:
		return DaemonSetDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}:
		return DeploymentDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}:
		return JobDeleteObj(cr, name)
	case schema.GroupVersionKind{Version: "v1", Kind: "Pod"}:
		return PodDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}:
		return StatefulSetDeleteObj(cr, name)
	case schema.GroupVersionKind{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}:
		return CsrDeleteObj(cr, name)
	}
	return false, fmt.Errorf("Resource not supported explicitly")
}

func DeleteFromSingleCR(c client.Client, cr *k8spluginv1alpha1.ResourceBundleState, name string, gvk schema.GroupVersionKind) error {

	found, _ := DeleteObj(cr, name, gvk)
	if found {
		fieldMgr := "emco-monitor"
		err := c.Status().Update(context.TODO(), cr, &client.UpdateOptions{FieldManager: fieldMgr})
		if err != nil {
			log.Printf("failed to update rbstate: %v\n", err)
			return err
		}
	}
	return nil
}

func DeleteFromAllCRs(c client.Client, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) error {
	var err error
	var found bool
	rbStatusList := &k8spluginv1alpha1.ResourceBundleStateList{}
	err = listResources(c, namespacedName.Namespace, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return fmt.Errorf("Did not find any CRs tracking this resource")
	}
	for _, cr := range rbStatusList.Items {
		orgStatus := cr.Status.DeepCopy()
		found, err = DeleteObj(&cr, namespacedName.Name, gvk)
		if found && err == nil {
			err = CommitCR(c, &cr, orgStatus)
		}
	}
	return err
}

func ClearLastApplied(annotations map[string]string) map[string]string {
	_, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]
	if ok {
		annotations["kubectl.kubernetes.io/last-applied-configuration"] = ""
	}
	return annotations
}

//Update CR status for generic resources
func UpdateResourceStatus(c client.Client, item *unstructured.Unstructured, name, namespace string) error {
	var err error

	rbStatusList, err := GetCRListForResource(c, item)
	if err != nil {
		return err
	}
	var found bool
	for _, cr := range rbStatusList.Items {
		orgStatus := cr.Status.DeepCopy()
		// Not scheduled for deletion
		if item.GetDeletionTimestamp() == nil {
			found, err = UpdateResourceStatusCR(&cr, item, name, namespace)

			if err == nil {
				err = CommitCR(c, &cr, orgStatus)
			}
		} else {
			found, err = DeleteResourceStatusCR(&cr, item.GetName(), item.GetNamespace(), item.GroupVersionKind())
			if found && err == nil {
				err = CommitCR(c, &cr, orgStatus)
			}
		}

	}
	return err
}

func DeleteResourceStatusFromAllCRs(c client.Client, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) error {
	var err error
	var found bool
	rbStatusList := &k8spluginv1alpha1.ResourceBundleStateList{}
	err = listResources(c, namespacedName.Namespace, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return fmt.Errorf("Did not find any CRs tracking this resource")
	}
	for _, cr := range rbStatusList.Items {
		orgStatus := cr.Status.DeepCopy()
		found, err = DeleteResourceStatusCR(&cr, namespacedName.Name, namespacedName.Namespace, gvk)
		if found && err == nil {
			err = CommitCR(c, &cr, orgStatus)
		}
	}
	return err
}

func DeleteResourceStatusCR(cr *k8spluginv1alpha1.ResourceBundleState, name, namespace string, gvk schema.GroupVersionKind) (bool, error) {
	var found bool
	length := len(cr.Status.ResourceStatuses)
	for i, rstatus := range cr.Status.ResourceStatuses {
		if (rstatus.Group == gvk.Group) && (rstatus.Version == gvk.Version) && (rstatus.Kind == gvk.Kind) && (rstatus.Name == name) && (rstatus.Namespace == namespace) {
			found = true
			//Delete that status from the array
			cr.Status.ResourceStatuses[i] = cr.Status.ResourceStatuses[length-1]
			cr.Status.ResourceStatuses[length-1] = k8spluginv1alpha1.ResourceStatus{}
			cr.Status.ResourceStatuses = cr.Status.ResourceStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func UpdateResourceStatusCR(cr *k8spluginv1alpha1.ResourceBundleState, item *unstructured.Unstructured, name, namespace string) (bool, error) {
	var found bool
	var res k8spluginv1alpha1.ResourceStatus

	// Clear up some fields to reduce size
	item.SetManagedFields([]metav1.ManagedFieldsEntry{})
	item.SetAnnotations(ClearLastApplied(item.GetAnnotations()))

	unstruct := item.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct, &res)
	if err != nil {
		log.Println("DefaultUnstructuredConverter error", name, namespace, err)
		return found, fmt.Errorf("Unknown resource")
	}

	group := item.GetObjectKind().GroupVersionKind().Group
	version := item.GetObjectKind().GroupVersionKind().Version
	kind := item.GetObjectKind().GroupVersionKind().Kind

	for _, rstatus := range cr.Status.ResourceStatuses {
		if (rstatus.Group == group) && (rstatus.Version == version) && (rstatus.Kind == kind) && (rstatus.Name == name) && (rstatus.Namespace == namespace) {
			found = true
			// Replace
			res.DeepCopyInto(&rstatus)
			break
		}
	}
	if !found {
		resBytes, err := json.Marshal(item)
		if err != nil {
			log.Println("json Marshal error for resource::", item, err)
			return found, err
		}
		// Add resource to ResourceMap
		res := k8spluginv1alpha1.ResourceStatus{
			Group:     group,
			Version:   version,
			Kind:      kind,
			Name:      name,
			Namespace: namespace,
			Res:       resBytes,
		}
		cr.Status.ResourceStatuses = append(cr.Status.ResourceStatuses, res)
	}
	return found, nil
}
