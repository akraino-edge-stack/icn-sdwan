// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package controllers

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"time"

	corev1 "k8s.io/api/core/v1"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/cnfprovider"
	"sdewan.akraino.org/sdewan/openwrt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var cnfCRNameSpace = "sdewan-system"
var cnfCRName = "cnf-status"
var inQueryStatus = false

// IStatusAction: defines the action to be executed based on CNF status
type IStatusAction interface {
	Execute(clientInfo *openwrt.OpenwrtClientInfo, status interface{}) error
}

// IpsecStatusAction: restart ipsec service if inactive
type IpsecStatusAction struct {
	client.Client
	Log logr.Logger
}

func (r *IpsecStatusAction) Execute(clientInfo *openwrt.OpenwrtClientInfo, status interface{}) error {
	stat := status.(map[string]interface{})
	val, ok := stat["InitConnection"]
	if !ok {
		return nil
	}

	if s := val.(string); s == "fail" {
		r.Log.Info("Restart IPSec service for " + clientInfo.Ip)
		openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
		service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
		_, err := service.ExecuteService("ipsec", "restart")
		if err != nil {
			r.Log.Info(err.Error())
			return err
		}
	}

	return nil
}

// SdewanCNFStatusController: query CNF status periodically
type SdewanCNFStatusController struct {
	client.Client
	Log           logr.Logger
	CheckInterval time.Duration
	actions       map[string]IStatusAction
	mux           sync.Mutex
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfstatuses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfstatuses/status,verbs=get;update;patch

func (r *SdewanCNFStatusController) SetupWithManager() error {
	r.actions = make(map[string]IStatusAction)
	r.RegisterAction("ipsec", &IpsecStatusAction{r.Client, r.Log})

	go wait.Until(r.SafeQuery, r.CheckInterval, wait.NeverStop)

	return nil
}

func (r *SdewanCNFStatusController) RegisterAction(module string, action IStatusAction) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.Log.Info("Register Action: " + module)
	if r.actions[module] == nil {
		r.actions[module] = action
	}
}

func (r *SdewanCNFStatusController) GetInstance(ctx context.Context) (*batchv1alpha1.CNFStatus, error) {
	instance := &batchv1alpha1.CNFStatus{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: cnfCRNameSpace,
		Name:      cnfCRName,
	}, instance)

	if errs.IsNotFound(err) {
		// No instance, create the instance
		r.Log.Info("Create New CNFStatus CR")
		instance = &batchv1alpha1.CNFStatus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cnfCRName,
				Namespace: cnfCRNameSpace,
			},
			Spec: batchv1alpha1.CNFStatusSpec{},
		}

		err = r.Create(ctx, instance)
		if err != nil {
			return &batchv1alpha1.CNFStatus{}, err
		}
	}

	return instance, nil
}

// Query CNFStatus information
func (r *SdewanCNFStatusController) SafeQuery() {
	doQuery := true
	r.mux.Lock()
	if !inQueryStatus {
		inQueryStatus = true
	} else {
		doQuery = false
	}
	r.mux.Unlock()

	if doQuery {
		r.query()

		r.mux.Lock()
		inQueryStatus = false
		r.mux.Unlock()
	}
}

func (r *SdewanCNFStatusController) query() {
	ctx := context.Background()

	instance, err := r.GetInstance(ctx)
	if err != nil {
		r.Log.Info(err.Error())
	} else {
		r.Log.Info("Query CNFStatus CR Instance: " + instance.ObjectMeta.Name)
	}

	// Set Status infomration
	instance.Status.AppliedGeneration = instance.Generation
	instance.Status.AppliedTime = &metav1.Time{Time: time.Now()}
	instance.Status.Information = []batchv1alpha1.CNFStatusInformation{}

	cnfPodList := &corev1.PodList{}
	err = r.List(ctx, cnfPodList, client.HasLabels{"sdewanPurpose"})
	if err != nil {
		r.Log.Info(err.Error())
		return
	}

	for _, cnfPod := range cnfPodList.Items {
		info := &batchv1alpha1.CNFStatusInformation{}
		info.Name = cnfPod.ObjectMeta.Name
		info.NameSpace = cnfPod.ObjectMeta.Namespace
		info.Node = cnfPod.Spec.NodeName
		info.Purpose = cnfPod.ObjectMeta.Labels["sdewanPurpose"]
		info.IP = cnfPod.Status.PodIP

		// Get CNF Status
		clientInfo := cnfprovider.CreateOpenwrtClient(cnfPod, r)
		openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
		status_client := openwrt.StatusClient{OpenwrtClient: openwrtClient}
		cnf_status, err := status_client.GetStatus()
		if err != nil {
			info.Status = "Not Available"
		} else {
			// ececute registered actions
			r.mux.Lock()
			var wg sync.WaitGroup
			for i, _ := range *cnf_status {
				if r.actions[(*cnf_status)[i].Name] != nil {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						err := r.actions[(*cnf_status)[index].Name].Execute(clientInfo, (*cnf_status)[index].Status)
						if err != nil {
							r.Log.Info(err.Error())
						}
					}(i)
				}
			}
			wg.Wait()
			r.mux.Unlock()

			p_data, _ := json.Marshal(cnf_status)
			info.Status = string(p_data)
		}
		instance.Status.Information = append(instance.Status.Information, *info)
	}

	// Update the CNFStatus CR
	err = r.Status().Update(ctx, instance)
	if err != nil {
		r.Log.Info(err.Error())
	}
}
