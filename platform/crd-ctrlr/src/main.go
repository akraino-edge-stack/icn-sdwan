// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = batchv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var checkInterval int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&checkInterval, "check-interval", 30,
		"The check interval of CRD Controller (seconds)")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "controller-leader-election-helper",
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Add indexer for rolebinding so that we can filter rolebindings by .subject
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.RoleBinding{}, ".subjects", func(rawObj runtime.Object) []string {
		var fieldValues []string
		rolebinding := rawObj.(*rbacv1.RoleBinding)
		for _, subject := range rolebinding.Subjects {
			if subject.Kind == "ServiceAccount" {
				fieldValues = append(fieldValues, fmt.Sprintf("system:serviceaccount:%s:%s", subject.Namespace, subject.Name))
			} else {
				fieldValues = append(fieldValues, subject.Name)
			}
		}
		return fieldValues
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &rbacv1.ClusterRoleBinding{}, ".subjects", func(rawObj runtime.Object) []string {
		var fieldValues []string
		clusterrolebinding := rawObj.(*rbacv1.ClusterRoleBinding)
		for _, subject := range clusterrolebinding.Subjects {
			if subject.Kind == "ServiceAccount" {
				fieldValues = append(fieldValues, fmt.Sprintf("system:serviceaccount:%s:%s", subject.Namespace, subject.Name))
			} else {
				fieldValues = append(fieldValues, subject.Name)
			}
		}
		return fieldValues
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "OwnBy", func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		var fieldValues []string
		pod := rawObj.(*corev1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil || owner.Kind != "ReplicaSet" {
			return nil
		}
		return append(fieldValues, owner.Name)
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.Mwan3PolicyReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Mwan3Policy"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Mwan3Policy")
		os.Exit(1)
	}
	if err = (&controllers.Mwan3RuleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Mwan3Rule"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Mwan3Rule")
		os.Exit(1)
	}
	if err = (&controllers.CNFNATReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CNFNAT"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CNFNAT")
		os.Exit(1)
	}
	if err = (&controllers.FirewallZoneReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FirewallZone"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FirewallZone")
		os.Exit(1)
	}
	if err = (&controllers.FirewallRuleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FirewallRule"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FirewallRule")
		os.Exit(1)
	}
	if err = (&controllers.FirewallSNATReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FirewallSNAT"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FirewallSNAT")
		os.Exit(1)
	}
	if err = (&controllers.FirewallDNATReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FirewallDNAT"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FirewallDNAT")
		os.Exit(1)
	}
	if err = (&controllers.FirewallForwardingReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FirewallForwarding"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FirewallForwarding")
		os.Exit(1)
	}

	if err = (&controllers.IpsecProposalReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IpsecProposal"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpsecProposal")
		os.Exit(1)
	}

	if err = (&controllers.IpsecHostReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IpsecHost"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpsecHost")
		os.Exit(1)
	}
	if err = batchv1alpha1.SetupBucketPermissionWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BucketPermission")
		os.Exit(1)
	}
	if err = batchv1alpha1.SetupLabelValidateWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "CNFLabelWebhook")
		os.Exit(1)
	}
	if err = (&controllers.IpsecSiteReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IpsecSite"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IpsecSite")
		os.Exit(1)
	}
	if err = (&controllers.CNFServiceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CNFService"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CNFService")
		os.Exit(1)
	}
	if err = (&controllers.SdewanApplicationReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SdewanApplication"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SdewanApplication")
		os.Exit(1)
	}
	if err = (&controllers.CNFRouteReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CNFRoute"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CNFRoute")
		os.Exit(1)
	}
	if err = (&controllers.CNFRouteRuleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CNFRouteRule"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CNFRouteRule")
		os.Exit(1)
	}
	if err = (&controllers.CNFLocalServiceReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("CNFLocalService"),
		CheckInterval: time.Duration(checkInterval) * time.Second,
		Scheme:        mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CNFLocalService")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("start CNFStatusController to query CNF status periodicly")
	if err = (&controllers.SdewanCNFStatusController{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("SdewanCNFStatus"),
		CheckInterval: time.Duration(checkInterval) * time.Second,
	}).SetupWithManager(); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SdewanCNFStatus")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
