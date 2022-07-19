// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	"encoding/json"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
)

type GvkElement struct {
	resource   string
	defaultRes bool
}

var defaultType GvkElement = GvkElement{defaultRes: true}

// List of resources with special handling in Monitor
var GVKList = map[schema.GroupVersionKind]GvkElement{
	{Version: "v1", Kind: "ConfigMap"}:                                               defaultType,
	{Group: "apps", Version: "v1", Kind: "DaemonSet"}:                                defaultType,
	{Group: "apps", Version: "v1", Kind: "Deployment"}:                               defaultType,
	{Group: "batch", Version: "v1", Kind: "Job"}:                                     defaultType,
	{Version: "v1", Kind: "Service"}:                                                 defaultType,
	{Version: "v1", Kind: "Pod"}:                                                     defaultType,
	{Group: "apps", Version: "v1", Kind: "StatefulSet"}:                              defaultType,
	{Group: "certificates.k8s.io", Version: "v1", Kind: "CertificateSigningRequest"}: defaultType,
	{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"}:                {resource: "roles", defaultRes: false},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"}:         {resource: "rolebindings", defaultRes: false},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}:         {resource: "roles", defaultRes: false},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}:  {resource: "clusterrolebindings", defaultRes: false},
	{Version: "v1", Kind: "ResourceQuota"}:                                           {resource: "resourcequotas", defaultRes: false},
	{Version: "v1", Kind: "Namespace"}:                                               {resource: "namespaces", defaultRes: false},
}

var GvkMap map[schema.GroupVersionKind]GvkElement

// ResourceBundleStateReconciler reconciles a ResourceBundleState object
type ControllerListReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	gvk    *schema.GroupVersionKind
}

func runtimeObjFromGVK(r schema.GroupVersionKind) runtime.Object {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r)
	return obj
}

//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/finalizers,verbs=update

func (r *ControllerListReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = slog.FromContext(ctx)

	log.Println("Reconcile", req.Name, req.Namespace, r.gvk)
	// Note: Create an unstructued type for the Client to use and set the Kind
	// so that it can GET/UPDATE/DELETE etc
	resource := &unstructured.Unstructured{}
	resource.SetGroupVersionKind(*r.gvk)

	err := r.Get(ctx, req.NamespacedName, resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			if g, ok := GvkMap[*r.gvk]; ok {
				var err1 error
				if g.defaultRes {
					err1 = DeleteFromAllCRs(r.Client, req.NamespacedName, *r.gvk)
				} else {
					err1 = DeleteResourceStatusFromAllCRs(r.Client, req.NamespacedName, *r.gvk)
				}
				return ctrl.Result{}, err1
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// If resource not a default resource for the controller
	// Add status to ResourceStatues array
	if g, ok := GvkMap[*r.gvk]; ok {
		if g.defaultRes {
			err = UpdateCR(r.Client, resource, req.NamespacedName, *r.gvk)
		} else {
			err = UpdateResourceStatus(r.Client, resource, req.NamespacedName.Name, req.NamespacedName.Namespace)
		}
	}
	if err != nil {
		// Requeue the update
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

//
func GetResourcesDynamically(group string, version string, resource string, namespace string) (
	[]unstructured.Unstructured, error) {

	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)
	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace("").
		List(ctx, metav1.ListOptions{})

	if err != nil {
		log.Println("Error listing resource:", resourceId, err)
		return nil, err
	}

	return list.Items, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControllerListReconciler) SetupWithManager(mgr ctrl.Manager) error {

	resource := &unstructured.Unstructured{}
	resource.SetGroupVersionKind(*r.gvk)

	return ctrl.NewControllerManagedBy(mgr).
		Named(r.gvk.Kind+"-emco-monitor").
		For(resource, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			labels := object.GetLabels()
			_, ok := labels["emco/deployment-id"]
			if !ok {
				return false
			}
			return true
		}))).
		Complete(r)
}

func SetupControllerForType(mgr ctrl.Manager, gv schema.GroupVersionKind, breakonError bool) error {
	r := ControllerListReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		gvk:    &gv,
	}

	if err := (&r).SetupWithManager(mgr); err != nil {
		log.Println(err, "unable to create controller", "controller", gv)
		if breakonError {
			os.Exit(1)
		}
	}
	return nil
}

type configResource struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Kind     string `json:"kind"`
	Resource string `json:"resource"`
}

// readConfigFile reads the specified smsConfig file to setup some env variables
func readGVKList(file string) ([]configResource, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Println("FILE not found::", file, err)
		return []configResource{}, err
	}
	defer f.Close()

	// Read the configuration from json file
	result := make([]configResource, 0)
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	//err = decoder.Decode(&conf)
	if err != nil {
		log.Println("FILE Decode error::", file, err)
		return []configResource{}, err
	}
	return result, nil
}

func ListCrds(group string, version string) ([]configResource, error) {
	result := []configResource{}

	config := ctrl.GetConfigOrDie()
	d := discovery.NewDiscoveryClientForConfigOrDie(config)
	resources, err := d.ServerResourcesForGroupVersion(group + "/" + version)
	if err != nil {
		return result, err
	}

	for _, res := range resources.APIResources {
		if !strings.Contains(res.Name, "/") {
			result = append(result, configResource{
				Group:    group,
				Version:  version,
				Kind:     res.Kind,
				Resource: res.Name,
			})
		}
	}
	return result, nil
}

func SetupControllers(mgr ctrl.Manager) error {

	// Copy map
	GvkMap = make(map[schema.GroupVersionKind]GvkElement, len(GVKList))
	for k, v := range GVKList {
		GvkMap[k] = v
	}

	l, err := readGVKList("/opt/emco/monitor/gvk.conf")
	if err != nil {
		log.Println("Error reading configmap for GVK List")
	}

	for _, gv := range l {
		if gv.Resource == "" {
			crds, err := ListCrds(gv.Group, gv.Version)
			if err != nil {
				log.Println("Invalid API group and version for the cluster")
				continue
			}

			for _, crd := range crds {
				gvk := schema.GroupVersionKind{
					Group:   crd.Group,
					Kind:    crd.Kind,
					Version: crd.Version,
				}
				if _, ok := GvkMap[gvk]; !ok {
					GvkMap[gvk] = GvkElement{defaultRes: false, resource: crd.Resource}
				}
			}
		} else {
			_, err = GetResourcesDynamically(gv.Group, gv.Version, gv.Resource, "default")
			if err != nil {
				log.Println("Invalid resource for the cluster", gv)
				continue
			}
			gvk := schema.GroupVersionKind{
				Group:   gv.Group,
				Kind:    gv.Kind,
				Version: gv.Version,
			}
			if _, ok := GvkMap[gvk]; !ok {
				GvkMap[gvk] = GvkElement{defaultRes: false, resource: gv.Resource}
			}
		}
	}
	log.Println("Adding controllers for::", GvkMap)
	for k, v := range GvkMap {
		SetupControllerForType(mgr, k, v.defaultRes)
	}
	return nil
}
