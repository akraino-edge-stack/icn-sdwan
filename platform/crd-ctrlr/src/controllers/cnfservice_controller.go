// SPDX-License-Identifier: Apache-2.0
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
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
)

// CNFServiceReconciler reconciles a CNFService object
type CNFServiceReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	mux           sync.Mutex
	CheckInterval time.Duration
}

var inSQueryStatus = false

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfservices/status,verbs=get;update;patch

func (r *CNFServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cnfservices", req.NamespacedName)
	during, _ := time.ParseDuration("5s")
	instance, err := r.GetInstance(req)
	if err != nil {
		if errs.IsNotFound(err) {
			// No instance
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{RequeueAfter: during}, nil
	}
	finalizerName := "cnfservice.finalizers.sdewan.akraino.org"
	delete_timestamp := getDeletionTempstamp(instance)
	log.Info("start Reconcile")
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
			log.Info("Added finalizer for CNFService")
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

func (r *CNFServiceReconciler) GetInstance(req ctrl.Request) (*batchv1alpha1.CNFService, error) {
	instance := &batchv1alpha1.CNFService{}
	err := r.Get(context.Background(), req.NamespacedName, instance)
	return instance, err
}

func (r *CNFServiceReconciler) processInstance(instance *batchv1alpha1.CNFService) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	// check service ip
	name := instance.Spec.FullName
	ipList, err := r.getIP4s(name)
	if err != nil || len(ipList) == 0 {
		if err != nil {
			r.removeNats(instance)
			r.Log.Error(err, "CNF Service")
		}
		return errors.New("Cannot reterive CNF Service ip")
	}

	// check  port
	svcPort := instance.Spec.Port
	if svcPort != "" {
		_, err = strconv.Atoi(svcPort)
		if err != nil {
			return errors.New("Port: " + err.Error())
		}
	}

	// check dport
	svcDPort := instance.Spec.DPort
	if svcDPort != "" {
		_, err = strconv.Atoi(svcDPort)
		if err != nil {
			return errors.New("svcDPort: " + err.Error())
		}
	}

	var curStatus = batchv1alpha1.CNFServiceStatus{
		SIp:   ipList[0],
		Port:  svcPort,
		DPort: svcDPort,
	}

	if !curStatus.IsEqual(&instance.Status) {
		r.removeNats(instance)
		r.addNats(instance, &curStatus)
		instance.Status = curStatus
		r.Log.Info("start update")
		r.Status().Update(context.Background(), instance)

	}

	return nil
}

func (r *CNFServiceReconciler) getIP4s(dns string) ([]string, error) {
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

func (r *CNFServiceReconciler) removeNats(instance *batchv1alpha1.CNFService) error {
	r.Log.Info("Deleting CNFNAT CR for CNF Service : " + instance.Name)
	nat_name := instance.Name + "nat"
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
	return nil
}

func (r *CNFServiceReconciler) addNats(instance *batchv1alpha1.CNFService, status *batchv1alpha1.CNFServiceStatus) error {
	r.Log.Info("Creating New CNFNAT CR for CNF Service : " + instance.Name)
	nat_name := instance.Name + "nat"
	nat_instance := &batchv1alpha1.CNFNAT{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nat_name,
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
		Spec: batchv1alpha1.CNFNATSpec{
			DestIp:   status.SIp,
			DestPort: status.Port,
			SrcDPort: status.DPort,
			Index:    "2",
			Proto:    "tcp",
			Target:   "DNAT",
		}}

	err := r.Create(context.Background(), nat_instance)
	if err != nil {
		r.Log.Error(err, "Creating NAT CR : "+nat_name)
	}
	return nil
}

func (r *CNFServiceReconciler) removeInstance(instance *batchv1alpha1.CNFService) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.removeNats(instance)
}

func (r *CNFServiceReconciler) check() {
	ls_list := &batchv1alpha1.CNFServiceList{}
	err := r.List(context.Background(), ls_list)
	r.Log.Info("start check")
	if err != nil {
		r.Log.Error(err, "Failed to list CNFService CRs")
	} else {
		if len(ls_list.Items) > 0 {
			for _, inst := range ls_list.Items {
				r.processInstance(&inst)
			}
		}
	}
}

func (r *CNFServiceReconciler) SafeCheck() {
	doCheck := true
	r.mux.Lock()
	if !inSQueryStatus {
		inSQueryStatus = true
	} else {
		doCheck = false
	}
	r.mux.Unlock()

	if doCheck {
		r.check()
		r.mux.Lock()
		inSQueryStatus = false
		r.mux.Unlock()
	}
}

func (r *CNFServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFService{}).
		Complete(r)
}
