/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"sdewan.akraino.org/sdewan/openwrt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var cnfCRNameSpace = "sdewan-system"
var cnfCRName = "cnf-status"
var inQueryStatus = false

// SdewanCNFStatusController: query CNF status periodically
type SdewanCNFStatusController struct {
	client.Client
	Log           logr.Logger
	CheckInterval time.Duration
	mux           sync.Mutex
}

func (r *SdewanCNFStatusController) SetupWithManager() error {
	go wait.Until(r.SafeQuery, r.CheckInterval, wait.NeverStop)

	return nil
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
	if inQueryStatus == false {
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
		clientInfo := &openwrt.OpenwrtClientInfo{Ip: info.IP, User: "root", Password: ""}
		openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
		status_client := openwrt.StatusClient{OpenwrtClient: openwrtClient}
		cnf_status, err := status_client.GetStatus()
		if err != nil {
			info.Status = "Not Available"
		} else {
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
