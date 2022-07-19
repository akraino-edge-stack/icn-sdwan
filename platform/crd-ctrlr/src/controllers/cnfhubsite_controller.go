// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package controllers

import (
	"context"
	"errors"
	"log"
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

var inHSQueryStatus = false

// CNFHubSiteReconciler reconciles a CNFHubSite object
type CNFHubSiteReconciler struct {
	client.Client
	Log           logr.Logger
	CheckInterval time.Duration
	Scheme        *runtime.Scheme
	mux           sync.Mutex
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfhubsites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfhubsites/status,verbs=get;update;patch

func (r *CNFHubSiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("CNFHubSite", req.NamespacedName)
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

	finalizerName := "cnfhubsite.finalizers.sdewan.akraino.org"
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
			log.Info("Added finalizer for CNFHubSite")
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

func (r *CNFHubSiteReconciler) getInstance(req ctrl.Request) (*batchv1alpha1.CNFHubSite, error) {
	instance := &batchv1alpha1.CNFHubSite{}
	err := r.Get(context.Background(), req.NamespacedName, instance)
	return instance, err
}

func (r *CNFHubSiteReconciler) getIP4s(dns string) ([]string, error) {
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

func (r *CNFHubSiteReconciler) processInstance(instance *batchv1alpha1.CNFHubSite) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	log.Println("Into hubsite processing")
	// check Type
	t := instance.Spec.Type
	log.Println("Type:", t)
	if t != "Hub" && t != "Device" {
		return errors.New("Invalid Type: should be Hub or Device")
	}

	ls := instance.Spec.Site
	sn := instance.Spec.Subnet

	if ls == "" && sn == "" {
		return errors.New("Invalid Site: neither of the url or the subnet set")
	}

	// check Site
	var lips []string
	if ls != "" {
		lips, err := r.getIP4s(ls)
		if err != nil || len(lips) == 0 {
			if err != nil {
				r.Log.Error(err, "Hub Site")
			}
			return errors.New("Cannot retrieve Site ip")
		}
	}

	// check subnet
	if sn != "" {
		_, _, err := net.ParseCIDR(sn)
		log.Println("Parsing subnet")
		if err != nil {
			r.Log.Error(err, "Subnet")
			return errors.New("Invalid Subnet")
		}
	}

	// check HubIP
	hip := instance.Spec.HubIP
	if hip == "" {
		if t == "Hub" {
			return errors.New("HubIP is required")
		}
	} else {
		ip := net.ParseIP(hip)
		if ip == nil {
			return errors.New("Invalid HubIP: " + hip)
		}
	}

	// check DevicePIP
	dpip := instance.Spec.DevicePIP
	if dpip == "" {
		if t == "Device" {
			return errors.New("DevicePIP is required")
		}
	} else {
		ip := net.ParseIP(dpip)
		if ip == nil {
			return errors.New("Invalid DevicePIP: " + dpip)
		}
	}

	var curStatus = batchv1alpha1.CNFHubSiteStatus{
		Type:      t,
		SiteIPs:   lips,
		Subnet:    sn,
		HubIP:     hip,
		DevicePIP: dpip,
		Message:   "",
	}

	if !curStatus.IsEqual(&instance.Status) {
		r.removeCRs(instance)
		r.addCRs(instance, &curStatus)
		instance.Status = curStatus
		r.Status().Update(context.Background(), instance)
	}

	return nil
}

func (r *CNFHubSiteReconciler) addCRs(instance *batchv1alpha1.CNFHubSite, status *batchv1alpha1.CNFHubSiteStatus) error {
	r.Log.Info("Creating New CRs for Hub Site : " + instance.Name)
	var dips []string
	if status.Subnet != "" {
		dips = append(dips, status.Subnet)
	}
	dips = append(dips, status.SiteIPs...)
	t := status.Type
	if t == "Hub" {
		// Create Route CR in Hub
		route_base_name := instance.Name + "route"
		for i, ip := range dips {
			route_name := route_base_name + strconv.Itoa(i)
			route_instance := &batchv1alpha1.CNFRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      route_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFRouteSpec{
					Dst:   ip,
					Dev:   "vti_" + status.HubIP,
					Table: "default",
				},
			}

			err := r.Create(context.Background(), route_instance)
			if err != nil {
				r.Log.Error(err, "Creating Route CR : "+route_name)
			}
		}
	} else {
		// Create Route CR and SNAT CR in Device
		route_base_name := instance.Name + "route"
		nat_base_name := instance.Name + "snat"
		for i, ip := range dips {
			// Route CR
			route_name := route_base_name + strconv.Itoa(i)
			route_instance := &batchv1alpha1.CNFRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      route_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFRouteSpec{
					Dst:   ip,
					Dev:   "#" + status.DevicePIP,
					Table: "cnf",
				},
			}
			log.Println(route_instance)

			err := r.Create(context.Background(), route_instance)
			if err != nil {
				r.Log.Error(err, "Creating Route CR : "+route_name)
			}

			// SNAT CR
			nat_name := nat_base_name + strconv.Itoa(i)
			nat_instance := &batchv1alpha1.CNFNAT{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nat_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFNATSpec{
					DestIp: ip,
					Dest:   "#source",
					SrcDIp: status.DevicePIP,
					Index:  "1",
					Target: "SNAT",
				},
			}

			err = r.Create(context.Background(), nat_instance)
			if err != nil {
				r.Log.Error(err, "Creating SNAT CR : "+nat_name)
			}
		}
	}

	return nil
}

func (r *CNFHubSiteReconciler) removeInstance(instance *batchv1alpha1.CNFHubSite) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.removeCRs(instance)
}

func (r *CNFHubSiteReconciler) removeCRs(instance *batchv1alpha1.CNFHubSite) error {
	r.Log.Info("Deleting CRs for Hub Site : " + instance.Name)
	var dips []string
	if instance.Status.Subnet != "" {
		dips = append(dips, instance.Status.Subnet)
	}
	dips = append(dips, instance.Status.SiteIPs...)

	t := instance.Status.Type
	if t == "Hub" {
		// Create Route CR in Hub
		route_base_name := instance.Name + "route"
		for i, _ := range dips {
			route_name := route_base_name + strconv.Itoa(i)
			route_instance := &batchv1alpha1.CNFRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      route_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFRouteSpec{},
			}

			err := r.Delete(context.Background(), route_instance)
			if err != nil {
				r.Log.Error(err, "Deleting Route CR : "+route_name)
			}

			// check resource
			err = wait.PollImmediate(time.Second, time.Second*10,
				func() (bool, error) {
					route_instance_temp := &batchv1alpha1.CNFRoute{}
					err_get := r.Get(context.Background(), client.ObjectKey{
						Namespace: instance.Namespace,
						Name:      route_name,
					}, route_instance_temp)

					if errs.IsNotFound(err_get) {
						return true, nil
					}
					r.Log.Info("Waiting for Deleting CR : " + route_name)
					return false, nil
				},
			)

			if err != nil {
				r.Log.Error(err, "Failed to delete CR : "+route_name)
			}
		}
	} else {
		// Delete Route CR and SNAT CR in Device
		route_base_name := instance.Name + "route"
		nat_base_name := instance.Name + "snat"
		for i, _ := range dips {
			// Route CR
			route_name := route_base_name + strconv.Itoa(i)
			route_instance := &batchv1alpha1.CNFRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      route_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFRouteSpec{},
			}

			err := r.Delete(context.Background(), route_instance)
			if err != nil {
				r.Log.Error(err, "Deleting Route CR : "+route_name)
			}

			// SNAT CR
			nat_name := nat_base_name + strconv.Itoa(i)
			nat_instance := &batchv1alpha1.CNFNAT{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nat_name,
					Namespace: instance.Namespace,
					Labels:    instance.Labels,
				},
				Spec: batchv1alpha1.CNFNATSpec{},
			}

			err = r.Delete(context.Background(), nat_instance)
			if err != nil {
				r.Log.Error(err, "Deleting SNAT CR : "+nat_name)
			}

			// check resource
			err = wait.PollImmediate(time.Second, time.Second*10,
				func() (bool, error) {
					route_instance_temp := &batchv1alpha1.CNFRoute{}
					err_get := r.Get(context.Background(), client.ObjectKey{
						Namespace: instance.Namespace,
						Name:      route_name,
					}, route_instance_temp)

					if errs.IsNotFound(err_get) {
						return true, nil
					}
					r.Log.Info("Waiting for Deleting CR : " + route_name)
					return false, nil
				},
			)

			if err != nil {
				r.Log.Error(err, "Failed to delete CR : "+route_name)
			}

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
	}

	return nil
}

func (r *CNFHubSiteReconciler) check() {
	ls_list := &batchv1alpha1.CNFHubSiteList{}
	err := r.List(context.Background(), ls_list)
	if err != nil {
		r.Log.Error(err, "Failed to list CNFHubSite CRs")
	} else {
		if len(ls_list.Items) > 0 {
			for _, inst := range ls_list.Items {
				r.Log.Info("Checking CNFHubSite: " + inst.Name)
				r.processInstance(&inst)
			}
		}
	}
}

// Regular check
func (r *CNFHubSiteReconciler) SafeCheck() {
	doCheck := true
	r.mux.Lock()
	if !inHSQueryStatus {
		inHSQueryStatus = true
	} else {
		doCheck = false
	}
	r.mux.Unlock()

	if doCheck {
		r.check()

		r.mux.Lock()
		inHSQueryStatus = false
		r.mux.Unlock()
	}
}

func (r *CNFHubSiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&batchv1alpha1.CNFHubSite{}, ps).
		Complete(r)
}
