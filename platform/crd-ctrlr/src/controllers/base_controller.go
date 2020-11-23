/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/basehandler"
	"sdewan.akraino.org/sdewan/cnfprovider"
	"sdewan.akraino.org/sdewan/openwrt"
)

// A global filter to catch the CNF deployments.
var Filter = builder.WithPredicates(predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		if _, ok := e.Meta.GetLabels()["sdewanPurpose"]; !ok {
			return false
		}
		return true
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		if _, ok := e.MetaOld.GetLabels()["sdewanPurpose"]; !ok {
			return false
		}
		pre_status := reflect.ValueOf(e.ObjectOld).Interface().(*appsv1.Deployment).Status
		post_status := reflect.ValueOf(e.ObjectNew).Interface().(*appsv1.Deployment).Status
		// Catch the available and unavailable pods of the Deployment
		if pre_status.AvailableReplicas != post_status.AvailableReplicas {
			return true
		}
		if pre_status.UnavailableReplicas != post_status.UnavailableReplicas {
			return true
		}
		return false
	},
})

// List the needed CR to specific events and return the reconcile Requests
func GetToRequestsFunc(r client.Client, crliststruct runtime.Object) func(h handler.MapObject) []reconcile.Request {

	return func(h handler.MapObject) []reconcile.Request {
		var enqueueRequest []reconcile.Request
		cnfName := h.Meta.GetLabels()["sdewanPurpose"]
		ctx := context.Background()
		r.List(ctx, crliststruct, client.MatchingLabels{"sdewanPurpose": cnfName})
		value := reflect.ValueOf(crliststruct)
		items := reflect.Indirect(value).FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			meta := items.Index(i).Field(1).Interface().(metav1.ObjectMeta)
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      meta.GetName(),
					Namespace: meta.GetNamespace(),
				}}
			enqueueRequest = append(enqueueRequest, req)

		}
		return enqueueRequest
	}
}

// A global filter to catch the change of cluster IP.
var IPFilter = builder.WithPredicates(predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		serviceName := e.MetaOld.GetName()
		if serviceName != "kubernetes" && serviceName != "istio-ingressgateway" {
			return false
		}
		pre_ip := reflect.ValueOf(e.ObjectOld).Interface().(*corev1.Service).Spec.ClusterIP
		post_ip := reflect.ValueOf(e.ObjectNew).Interface().(*corev1.Service).Spec.ClusterIP
		return pre_ip != post_ip
	},
})

// List the needed CR to specific events and return the reconcile Requests
func GetServiceToRequestsFunc(r client.Client) func(h handler.MapObject) []reconcile.Request {

	return func(h handler.MapObject) []reconcile.Request {
		var cnfName string
		deploymentList := &appsv1.DeploymentList{}
		podList := &corev1.PodList{}
		ctx := context.Background()
		r.List(ctx, deploymentList)
		for _, deployment := range deploymentList.Items {
			if _, ok := deployment.ObjectMeta.GetLabels()["sdewanPurpose"]; !ok {
				continue
			}
			cnfName = deployment.ObjectMeta.GetLabels()["sdewanPurpose"]
			// Break there to fetch the first deployment with sdewanPurpose label.
			// TODO: For multi-namespace deployments with different sdewanPurpose label, a cnfName list is needed to hold them.
			break
		}
		r.List(ctx, podList, client.MatchingLabels{"sdewanPurpose": cnfName})
		for _, pod := range podList.Items {
			clientInfo := &openwrt.OpenwrtClientInfo{Ip: pod.Status.PodIP, User: "root", Password: ""}
			openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
			service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
			service.ExecuteService("firewall", "restart")
		}
		return []reconcile.Request{}
	}
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func getPurpose(instance runtime.Object) string {
	value := reflect.ValueOf(instance)
	field := reflect.Indirect(value).FieldByName("Labels")
	labels := field.Interface().(map[string]string)
	return labels["sdewanPurpose"]
}

func getDeletionTempstamp(instance runtime.Object) *metav1.Time {
	value := reflect.ValueOf(instance)
	field := reflect.Indirect(value).FieldByName("DeletionTimestamp")
	return field.Interface().(*metav1.Time)
}

func getFinalizers(instance runtime.Object) []string {
	value := reflect.ValueOf(instance)
	field := reflect.Indirect(value).FieldByName("Finalizers")
	return field.Interface().([]string)
}

func setStatus(instance runtime.Object, status batchv1alpha1.SdewanStatus) {
	value := reflect.ValueOf(instance)
	field_status := reflect.Indirect(value).FieldByName("Status")
	if status.State == batchv1alpha1.InSync {
		field_gv := reflect.Indirect(value).FieldByName("Generation")
		status.AppliedGeneration = field_gv.Interface().(int64)
		status.AppliedTime = &metav1.Time{Time: time.Now()}
		status.Message = ""
	} else {
		status.AppliedGeneration = 0
		status.AppliedTime = nil
	}
	field_status.Set(reflect.ValueOf(status))
}

func appendFinalizer(instance runtime.Object, item string) {
	value := reflect.ValueOf(instance)
	field := reflect.Indirect(value).FieldByName("ObjectMeta")
	base_obj := field.Interface().(metav1.ObjectMeta)
	base_obj.Finalizers = append(base_obj.Finalizers, item)
	field.Set(reflect.ValueOf(base_obj))
}

func removeFinalizer(instance runtime.Object, item string) {
	value := reflect.ValueOf(instance)
	field := reflect.Indirect(value).FieldByName("ObjectMeta")
	base_obj := field.Interface().(metav1.ObjectMeta)
	base_obj.Finalizers = removeString(base_obj.Finalizers, item)
	field.Set(reflect.ValueOf(base_obj))
}

func net2iface(net string, deployment appsv1.Deployment) (string, error) {
	type Iface struct {
		DefaultGateway bool `json:"defaultGateway,string"`
		Interface      string
		Name           string
	}
	type NfnNet struct {
		Type      string
		Interface []Iface
	}
	ann := deployment.Spec.Template.Annotations
	nfnNet := NfnNet{}
	err := json.Unmarshal([]byte(ann["k8s.plugin.opnfv.org/nfn-network"]), &nfnNet)
	if err != nil {
		return "", err
	}
	for _, iface := range nfnNet.Interface {
		if iface.Name == net {
			return iface.Interface, nil
		}
	}
	return "", errors.New(fmt.Sprintf("No matched network in annotation: %s", net))
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;list;watch

// Common Reconcile Processing
func ProcessReconcile(r client.Client, logger logr.Logger, req ctrl.Request, handler basehandler.ISdewanHandler) (ctrl.Result, error) {
	ctx := context.Background()
	log := logger.WithValues(handler.GetType(), req.NamespacedName)
	during, _ := time.ParseDuration("5s")

	instance, err := handler.GetInstance(r, ctx, req)
	if err != nil {
		if errs.IsNotFound(err) {
			// No instance
			return ctrl.Result{}, nil
		}

		err1, ok := err.(*AppCRError)
		if ok && err1.Code == 404 {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{RequeueAfter: during}, nil
	}
	purpose := getPurpose(instance)
	cnf, err := cnfprovider.NewOpenWrt(req.NamespacedName.Namespace, purpose, r)
	if err != nil {
		log.Error(err, "Failed to get cnf")
		setStatus(instance, batchv1alpha1.SdewanStatus{State: batchv1alpha1.Unknown, Message: err.Error()})
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update status for "+handler.GetType())
			return ctrl.Result{}, err
		}
		// A new event are supposed to be received upon cnf ready
		// so not requeue
		return ctrl.Result{}, nil
	}
	finalizerName := handler.GetFinalizer()
	delete_timestamp := getDeletionTempstamp(instance)

	if delete_timestamp.IsZero() {
		// creating or updating CR
		if cnf == nil {
			// no cnf exists
			log.Info("No cnf exist, so not create/update " + handler.GetType())
			return ctrl.Result{}, nil
		}
		changed, err := cnf.AddOrUpdateObject(handler, instance)
		if err != nil {
			log.Error(err, "Failed to add/update "+handler.GetType())
			setStatus(instance, batchv1alpha1.SdewanStatus{State: batchv1alpha1.Applying, Message: err.Error()})
			err = r.Status().Update(ctx, instance)
			if err != nil {
				log.Error(err, "Failed to update status for "+handler.GetType())
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: during}, nil
		}
		finalizers := getFinalizers(instance)
		if !containsString(finalizers, finalizerName) {
			appendFinalizer(instance, finalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Added finalizer for " + handler.GetType())
		}
		if changed {
			setStatus(instance, batchv1alpha1.SdewanStatus{State: batchv1alpha1.InSync})

			err = r.Status().Update(ctx, instance)
			if err != nil {
				log.Error(err, "Failed to update status for "+handler.GetType())
				return ctrl.Result{}, err
			}
		}
	} else {
		// deletin CR
		if cnf == nil {
			// no cnf exists
			finalizers := getFinalizers(instance)
			if containsString(finalizers, finalizerName) {
				// instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
				removeFinalizer(instance, finalizerName)
				if err := r.Update(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		_, err := cnf.DeleteObject(handler, instance)

		if err != nil {
			err2, ok := err.(*openwrt.OpenwrtError)
			if !ok || err2.Code != 404 {
				log.Error(err, "Failed to delete "+handler.GetType())
				setStatus(instance, batchv1alpha1.SdewanStatus{State: batchv1alpha1.Deleting, Message: err.Error()})
				err = r.Status().Update(ctx, instance)
				if err != nil {
					log.Error(err, "Failed to update status for "+handler.GetType())
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: during}, nil
			}
		}
		finalizers := getFinalizers(instance)
		if containsString(finalizers, finalizerName) {
			removeFinalizer(instance, finalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}
