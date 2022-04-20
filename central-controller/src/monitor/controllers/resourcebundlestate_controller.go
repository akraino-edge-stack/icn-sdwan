// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"fmt"
)

// ResourceBundleStateReconciler reconciles a ResourceBundleState object
type ResourceBundleStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceBundleState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ResourceBundleStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = slog.FromContext(ctx)
	log.Println("Reconcile CR", req)
	rbstate := &k8spluginv1alpha1.ResourceBundleState{}
	err := r.Get(context.TODO(), req.NamespacedName, rbstate)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Object not found: %+v. Ignore as it must have been deleted.\n", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Printf("Failed to get object: %+v\n", req.NamespacedName)
		return ctrl.Result{}, err
	}
	rbstate.Status.Ready = true
	orgStatus := &k8spluginv1alpha1.ResourceBundleStateStatus{}
	rbstate.Status.DeepCopyInto(orgStatus)

	err = r.updatePods(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding podstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateServices(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding servicestatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateConfigMaps(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding configmapstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateDeployments(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding deploymentstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateDaemonSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding daemonSetstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateJobs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding jobstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateStatefulSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding statefulSetstatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateCsrs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding csrStatuses: %v\n", err)
		return ctrl.Result{}, err
	}

	err = r.updateDynResources(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding dynamic resources: %v\n", err)
		return ctrl.Result{}, err
	}

	fmt.Println("Commit CR with status")
	fmt.Println(rbstate)
	err = CommitCR(r.Client, rbstate, orgStatus)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceBundleStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8spluginv1alpha1.ResourceBundleState{}).
		Complete(r)
}

func (r *ResourceBundleStateReconciler) updateServices(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	serviceList := &corev1.ServiceList{}

	err := listResources(r.Client, rbstate.Namespace, selectors, serviceList)
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		return err
	}

	for _, svc := range serviceList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&svc)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			ServiceUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updatePods(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the pods tracked
	podList := &corev1.PodList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return err
	}

	for _, pod := range podList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			PodUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}
	return nil
}

func (r *ResourceBundleStateReconciler) updateConfigMaps(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the ConfigMaps created as well
	configMapList := &corev1.ConfigMapList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, configMapList)
	if err != nil {
		log.Printf("Failed to list configMaps: %v", err)
		return err
	}
	for _, cm := range configMapList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cm)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			ConfigMapUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateDeployments(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Deployments created as well
	deploymentList := &appsv1.DeploymentList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, deploymentList)
	if err != nil {
		log.Printf("Failed to list deployments: %v", err)
		return err
	}

	for _, dep := range deploymentList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dep)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			DeploymentUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}
	return nil
}

func (r *ResourceBundleStateReconciler) updateDaemonSets(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the DaemonSets created as well
	daemonSetList := &appsv1.DaemonSetList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, daemonSetList)
	if err != nil {
		log.Printf("Failed to list DaemonSets: %v", err)
		return err
	}

	for _, ds := range daemonSetList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ds)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			DaemonSetUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateJobs(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	jobList := &v1.JobList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, jobList)
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
		return err
	}

	for _, job := range jobList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&job)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			JobUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateStatefulSets(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the StatefulSets created as well
	statefulSetList := &appsv1.StatefulSetList{}
	err := listResources(r.Client, rbstate.Namespace, selectors, statefulSetList)
	if err != nil {
		log.Printf("Failed to list statefulSets: %v", err)
		return err
	}

	for _, sfs := range statefulSetList.Items {
		newUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&sfs)
		if err == nil {
			un := &unstructured.Unstructured{}
			un.SetUnstructuredContent(newUnstr)
			StatefulSetUpdateStatus(rbstate, un)
		} else {
			return err
		}
	}

	return nil
}

func (r *ResourceBundleStateReconciler) updateCsrs(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the csrs tracked
	csrList := &certsapi.CertificateSigningRequestList{}
	err := listResources(r.Client, "", selectors, csrList)
	if err != nil {
		log.Printf("Failed to list csrs: %v", err)
		return err
	}

	rbstate.Status.CsrStatuses = []certsapi.CertificateSigningRequest{}

	for _, csr := range csrList.Items {
		resStatus := certsapi.CertificateSigningRequest{
			TypeMeta:   csr.TypeMeta,
			ObjectMeta: csr.ObjectMeta,
			Status:     csr.Status,
			Spec:       csr.Spec,
		}
		resStatus.ObjectMeta.ManagedFields = []metav1.ManagedFieldsEntry{}
		resStatus.Annotations = ClearLastApplied(resStatus.Annotations)
		rbstate.Status.CsrStatuses = append(rbstate.Status.CsrStatuses, resStatus)
	}
	return nil
}

// updateDynResources updates non default resources
func (r *ResourceBundleStateReconciler) updateDynResources(rbstate *k8spluginv1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	for gvk, item := range GvkMap {
		if item.defaultRes {
			// Already handled
			continue
		}
		resourceList := &unstructured.UnstructuredList{}
		resourceList.SetGroupVersionKind(gvk)

		err := listResources(r.Client, rbstate.Namespace, selectors, resourceList)
		if err != nil {
			log.Printf("Failed to list resources: %v", err)
			return err
		}
		for _, res := range resourceList.Items {
			err = UpdateResourceStatus(r.Client, &res, res.GetName(), res.GetNamespace())
			if err != nil {
				log.Println("Error updating status for resource", gvk, res.GetName())
			}
		}
	}
	return nil
}
