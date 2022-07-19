// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package controllers

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
)

var inLSQueryStatus = false

// CNFLocalServiceReconciler reconciles a CNFLocalService object
type CNFLocalServiceReconciler struct {
	client.Client
	Log           logr.Logger
	CheckInterval time.Duration
	Scheme        *runtime.Scheme
	mux           sync.Mutex
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnflocalservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnflocalservices/status,verbs=get;update;patch

func (r *CNFLocalServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("CNFLocalService", req.NamespacedName)
	during, _ := time.ParseDuration("5s")

	instance, err := r.getInstance(req)
	if err != nil {
		if errs.IsNotFound(err) {
			// No instance
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{RequeueAfter: during}, nil
	}

	finalizerName := "cnflocalservice.finalizers.sdewan.akraino.org"
	delete_timestamp := getDeletionTempstamp(instance)

	if delete_timestamp.IsZero() {
		// Creating or updating CR
		// Process instance
		err = r.processInstance(instance)
		if err != nil {
			log.Error(err, "Adding/Updating CR")
			instance.Status.Message = err.Error()
			r.Status().Update(ctx, instance)

			return ctrl.Result{}, err
		}

		finalizers := getFinalizers(instance)
		if !containsString(finalizers, finalizerName) {
			appendFinalizer(instance, finalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Added finalizer for CNFLocalService")
		}
	} else {
		// Deleting CR
		// Remove instance
		err = r.removeInstance(instance)
		if err != nil {
			log.Error(err, "Deleting CR")
			return ctrl.Result{RequeueAfter: during}, nil
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

func (r *CNFLocalServiceReconciler) getInstance(req ctrl.Request) (*batchv1alpha1.CNFLocalService, error) {
	instance := &batchv1alpha1.CNFLocalService{}
	err := r.Get(context.Background(), req.NamespacedName, instance)
	return instance, err
}

func (r *CNFLocalServiceReconciler) getIP4s(dns string) ([]string, error) {
	ips, err := net.LookupIP(dns)
	var ip4s []string

	if err == nil {
		for _, ip := range ips {
			if strings.Contains(ip.String(), ".") {
				ip4s = append(ip4s, ip.String())
			}
		}
	}

	return ip4s, err
}

func (r *CNFLocalServiceReconciler) processInstance(instance *batchv1alpha1.CNFLocalService) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	// check local service
	ls := instance.Spec.LocalService
	lips, err := r.getIP4s(ls)
	if err != nil || len(lips) == 0 {
		if err != nil {
			r.Log.Error(err, "Local Service")
		}
		return errors.New("Cannot reterive LocalService ip")
	}

	// check remote service
	rs := instance.Spec.RemoteService
	rips, err := r.getIP4s(rs)
	if err != nil || len(rips) == 0 {
		if err != nil {
			r.Log.Error(err, "Remote Service")
		}
		return errors.New("Cannot reterive RemoteService ip")
	}

	// check local port
	lp := instance.Spec.LocalPort
	if lp != "" {
		_, err = strconv.Atoi(lp)
		if err != nil {
			return errors.New("LocalPort: " + err.Error())
		}
	}

	// check remote port
	rp := instance.Spec.RemotePort
	if rp != "" {
		_, err = strconv.Atoi(rp)
		if err != nil {
			return errors.New("RemotePort: " + err.Error())
		}
	}

	var curStatus = batchv1alpha1.CNFLocalServiceStatus{
		LocalIP:    lips[0],
		LocalPort:  lp,
		RemoteIPs:  rips,
		RemotePort: rp,
		Message:    "",
	}

	if !curStatus.IsEqual(&instance.Status) {
		r.removeNats(instance)
		r.addNats(instance, &curStatus)
		instance.Status = curStatus
		r.Status().Update(context.Background(), instance)
	}

	return nil
}

func (r *CNFLocalServiceReconciler) addNats(instance *batchv1alpha1.CNFLocalService, status *batchv1alpha1.CNFLocalServiceStatus) error {
	r.Log.Info("Creating New CNFNAT CR for Local Service : " + instance.Name)
	nat_base_name := instance.Name + "nat"
	for i, rip := range status.RemoteIPs {
		nat_name := nat_base_name + strconv.Itoa(i)
		nat_instance := &batchv1alpha1.CNFNAT{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nat_name,
				Namespace: instance.Namespace,
				Labels:    instance.Labels,
			},
			Spec: batchv1alpha1.CNFNATSpec{
				SrcDIp:   rip,
				SrcDPort: status.RemotePort,
				DestIp:   status.LocalIP,
				DestPort: status.LocalPort,
				Proto:    "tcp",
				Target:   "DNAT",
			},
		}

		err := r.Create(context.Background(), nat_instance)
		if err != nil {
			r.Log.Error(err, "Creating NAT CR : "+nat_name)
		}
	}
	return nil
}

func (r *CNFLocalServiceReconciler) removeInstance(instance *batchv1alpha1.CNFLocalService) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.removeNats(instance)
}

func (r *CNFLocalServiceReconciler) removeNats(instance *batchv1alpha1.CNFLocalService) error {
	r.Log.Info("Deleting CNFNAT CR for Local Service : " + instance.Name)
	nat_base_name := instance.Name + "nat"
	for i, _ := range instance.Status.RemoteIPs {
		nat_name := nat_base_name + strconv.Itoa(i)
		nat_instance := &batchv1alpha1.CNFNAT{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nat_name,
				Namespace: instance.Namespace,
				Labels:    instance.Labels,
			},
			Spec: batchv1alpha1.CNFNATSpec{},
		}

		err := r.Delete(context.Background(), nat_instance)
		if err != nil {
			r.Log.Error(err, "Deleting NAT CR : "+nat_name)
		}

		// check resource
		err = wait.PollImmediate(time.Second, time.Second*10,
			func() (bool, error) {
				nat_instance_temp := &batchv1alpha1.CNFNAT{}
				err_get := r.Get(context.Background(), client.ObjectKey{
					Namespace: instance.Namespace,
					Name:      nat_name,
				}, nat_instance_temp)

				if errs.IsNotFound(err_get) {
					return true, nil
				}
				r.Log.Info("Waiting for Deleting CR : " + nat_name)
				return false, nil
			},
		)

		if err != nil {
			r.Log.Error(err, "Failed to delete CR : "+nat_name)
		}
	}

	return nil
}

// Query CNFStatus information
func (r *CNFLocalServiceReconciler) check() {
	ls_list := &batchv1alpha1.CNFLocalServiceList{}
	err := r.List(context.Background(), ls_list)
	if err != nil {
		r.Log.Error(err, "Failed to list CNFLocalService CRs")
	} else {
		if len(ls_list.Items) > 0 {
			for _, inst := range ls_list.Items {
				r.Log.Info("Checking CNFLocalService: " + inst.Name)
				r.processInstance(&inst)
			}
		}
	}
}

// Query CNFStatus information
func (r *CNFLocalServiceReconciler) SafeCheck() {
	doCheck := true
	r.mux.Lock()
	if !inLSQueryStatus {
		inLSQueryStatus = true
	} else {
		doCheck = false
	}
	r.mux.Unlock()

	if doCheck {
		r.check()

		r.mux.Lock()
		inLSQueryStatus = false
		r.mux.Unlock()
	}
}

func (r *CNFLocalServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Start the loop to check ip address change of local/remote services
	go func() {
		interval := time.After(r.CheckInterval)
		for {
			select {
			case <-interval:
				r.SafeCheck()
				interval = time.After(r.CheckInterval)
			case <-context.Background().Done():
				return
			}
		}
	}()

	ps := builder.WithPredicates(predicate.GenerationChangedPredicate{})
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFLocalService{}, ps).
		Complete(r)
}
