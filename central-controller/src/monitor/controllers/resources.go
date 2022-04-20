// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"fmt"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func ServiceUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {
	var found bool
	var service v1.Service
	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &service)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	for i, rstatus := range cr.Status.ServiceStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == service.Name {
			service.Status.DeepCopyInto(&cr.Status.ServiceStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		svc := v1.Service{
			TypeMeta:   service.TypeMeta,
			ObjectMeta: service.ObjectMeta,
			Status:     service.Status,
			Spec:       service.Spec,
		}
		svc.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		svc.ObjectMeta.Annotations = ClearLastApplied(svc.Annotations)
		cr.Status.ServiceStatuses = append(cr.Status.ServiceStatuses, svc)
	}
	return found, nil
}

func ServiceDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.ServiceStatuses)
	for i, rstatus := range cr.Status.ServiceStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.ServiceStatuses[i] = cr.Status.ServiceStatuses[length-1]
			cr.Status.ServiceStatuses[length-1] = v1.Service{}
			cr.Status.ServiceStatuses = cr.Status.ServiceStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func ConfigMapDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.ConfigMapStatuses)
	for i, rstatus := range cr.Status.ConfigMapStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.ConfigMapStatuses[i] = cr.Status.ConfigMapStatuses[length-1]
			cr.Status.ConfigMapStatuses = cr.Status.ConfigMapStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func ConfigMapUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var cm v1.ConfigMap
	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &cm)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for _, rstatus := range cr.Status.ConfigMapStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == cm.Name {
			found = true
			break
		}
	}

	if !found {
		// Add it to CR
		c := v1.ConfigMap{
			TypeMeta:   cm.TypeMeta,
			ObjectMeta: cm.ObjectMeta,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.ConfigMapStatuses = append(cr.Status.ConfigMapStatuses, c)
	}

	return found, nil
}

func DaemonSetDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.DaemonSetStatuses)
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.DaemonSetStatuses[i] = cr.Status.DaemonSetStatuses[length-1]
			cr.Status.DaemonSetStatuses[length-1].Status = appsv1.DaemonSetStatus{}
			cr.Status.DaemonSetStatuses = cr.Status.DaemonSetStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func DaemonSetUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var ds appsv1.DaemonSet
	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &ds)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ds.Name {
			ds.Status.DeepCopyInto(&cr.Status.DaemonSetStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := appsv1.DaemonSet{
			TypeMeta:   ds.TypeMeta,
			ObjectMeta: ds.ObjectMeta,
			Spec:       ds.Spec,
			Status:     ds.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.DaemonSetStatuses = append(cr.Status.DaemonSetStatuses, c)
	}
	return found, nil
}

func DeploymentDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.DeploymentStatuses)
	for i, rstatus := range cr.Status.DeploymentStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.DeploymentStatuses[i] = cr.Status.DeploymentStatuses[length-1]
			cr.Status.DeploymentStatuses[length-1].Status = appsv1.DeploymentStatus{}
			cr.Status.DeploymentStatuses = cr.Status.DeploymentStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func DeploymentUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var dm appsv1.Deployment
	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &dm)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DeploymentStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == dm.Name {
			dm.Status.DeepCopyInto(&cr.Status.DeploymentStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := appsv1.Deployment{
			TypeMeta:   dm.TypeMeta,
			ObjectMeta: dm.ObjectMeta,
			Spec:       dm.Spec,
			Status:     dm.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.DeploymentStatuses = append(cr.Status.DeploymentStatuses, c)
	}
	return found, nil
}

func JobDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.JobStatuses)
	for i, rstatus := range cr.Status.JobStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.JobStatuses[i] = cr.Status.JobStatuses[length-1]
			cr.Status.JobStatuses[length-1].Status = batchv1.JobStatus{}
			cr.Status.JobStatuses = cr.Status.JobStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func JobUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var job batchv1.Job
	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &job)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.JobStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == job.Name {
			job.Status.DeepCopyInto(&cr.Status.JobStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := batchv1.Job{
			TypeMeta:   job.TypeMeta,
			ObjectMeta: job.ObjectMeta,
			Spec:       job.Spec,
			Status:     job.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.JobStatuses = append(cr.Status.JobStatuses, c)
	}
	return found, nil
}

func PodUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {
	var found bool

	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	var pod v1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &pod)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	for i, rstatus := range cr.Status.PodStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == pod.Name {
			pod.Status.DeepCopyInto(&cr.Status.PodStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		ps := v1.Pod{
			TypeMeta:   pod.TypeMeta,
			ObjectMeta: pod.ObjectMeta,
			Status:     pod.Status,
			Spec:       pod.Spec,
		}
		ps.ObjectMeta = metav1.ObjectMeta{}
		ps.ObjectMeta.Name = pod.GetName()
		ps.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		ps.Annotations = ClearLastApplied(ps.Annotations)
		cr.Status.PodStatuses = append(cr.Status.PodStatuses, ps)
	}
	return found, nil
}

func PodDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.PodStatuses)
	for i, rstatus := range cr.Status.PodStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.PodStatuses[i] = cr.Status.PodStatuses[length-1]
			cr.Status.PodStatuses[length-1] = v1.Pod{}
			cr.Status.PodStatuses = cr.Status.PodStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func StatefulSetDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.StatefulSetStatuses)
	for i, rstatus := range cr.Status.StatefulSetStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.StatefulSetStatuses[i] = cr.Status.StatefulSetStatuses[length-1]
			cr.Status.StatefulSetStatuses[length-1].Status = appsv1.StatefulSetStatus{}
			cr.Status.StatefulSetStatuses = cr.Status.StatefulSetStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func StatefulSetUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var ss appsv1.StatefulSet

	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &ss)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.StatefulSetStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ss.Name {
			ss.Status.DeepCopyInto(&cr.Status.StatefulSetStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		c := appsv1.StatefulSet{
			TypeMeta:   ss.TypeMeta,
			ObjectMeta: ss.ObjectMeta,
			Spec:       ss.Spec,
			Status:     ss.Status,
		}
		c.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		c.Annotations = ClearLastApplied(c.Annotations)
		cr.Status.StatefulSetStatuses = append(cr.Status.StatefulSetStatuses, c)
	}
	return found, nil
}

func CsrDeleteObj(cr *k8spluginv1alpha1.ResourceBundleState, name string) (bool, error) {
	var found bool
	length := len(cr.Status.CsrStatuses)
	for i, rstatus := range cr.Status.CsrStatuses {
		if rstatus.Name == name {
			found = true
			//Delete that status from the array
			cr.Status.CsrStatuses[i] = cr.Status.CsrStatuses[length-1]
			cr.Status.CsrStatuses[length-1] = certsapi.CertificateSigningRequest{}
			cr.Status.CsrStatuses = cr.Status.CsrStatuses[:length-1]
			break
		}
	}
	return found, nil
}

func CsrUpdateStatus(cr *k8spluginv1alpha1.ResourceBundleState, obj *unstructured.Unstructured) (bool, error) {

	var found bool
	var csr *certsapi.CertificateSigningRequest

	// Convert the unstructured object to actual
	unstructured := obj.UnstructuredContent()
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &csr)
	if err != nil {
		return found, fmt.Errorf("Unknown resource %v", obj)
	}
	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.CsrStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == csr.Name {
			csr.Status.DeepCopyInto(&cr.Status.CsrStatuses[i].Status)
			found = true
			break
		}
	}
	if !found {
		// Add it to CR
		cr.Status.CsrStatuses = append(cr.Status.CsrStatuses, certsapi.CertificateSigningRequest{
			TypeMeta:   csr.TypeMeta,
			ObjectMeta: csr.ObjectMeta,
			Status:     csr.Status,
			Spec:       csr.Spec,
		})
	}
	return found, nil
}
