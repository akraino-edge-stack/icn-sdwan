package cnfprovider

import (
	"context"
	"errors"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	basehandler "sdewan.akraino.org/sdewan/basehandler"
	"sdewan.akraino.org/sdewan/openwrt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("OpenWrtProvider")

type OpenWrtProvider struct {
	Namespace     string
	SdewanPurpose string
	Deployment    appsv1.Deployment
	K8sClient     client.Client
}

func NewOpenWrt(namespace string, sdewanPurpose string, k8sClient client.Client) (*OpenWrtProvider, error) {
	ctx := context.Background()
	deployments := &appsv1.DeploymentList{}
	err := k8sClient.List(ctx, deployments, client.MatchingLabels{"sdewanPurpose": sdewanPurpose})
	if err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	if len(deployments.Items) > 1 {
		return nil, errors.New("More than one deployment exists")
	}
	if len(deployments.Items) < 1 {
		// return (nil, nil) to indicate that no cnf exists
		return nil, nil
	}
	return &OpenWrtProvider{namespace, sdewanPurpose, deployments.Items[0], k8sClient}, nil
}

func (p *OpenWrtProvider) AddOrUpdateObject(handler basehandler.ISdewanHandler, instance runtime.Object) (bool, error) {
	reqLogger := log.WithValues(handler.GetType(), handler.GetName(instance), "cnf", p.Deployment.Name)
	ctx := context.Background()
	ReplicaSetList := &appsv1.ReplicaSetList{}
	err := p.K8sClient.List(ctx, ReplicaSetList, client.MatchingLabels{"sdewanPurpose": p.SdewanPurpose})
	if err != nil {
		return false, err
	}
	if len(ReplicaSetList.Items) != 1 {
		return false, errors.New(fmt.Sprintf("More than one of repicaset exist with label: sdewanPurpose=%s", p.SdewanPurpose))
	}
	podList := &corev1.PodList{}
	err = p.K8sClient.List(ctx, podList, client.MatchingFields{"OwnBy": ReplicaSetList.Items[0].ObjectMeta.Name})
	if err != nil {
		return false, err
	}
	new_instance, err := handler.Convert(instance, p.Deployment)
	if err != nil {
		return false, err
	}
	cnfChanged := false
	for _, pod := range podList.Items {
		clientInfo := &openwrt.OpenwrtClientInfo{Ip: pod.Status.PodIP, User: "root", Password: ""}
		runtime_instance, err := handler.GetObject(clientInfo, new_instance.GetName())
		changed := false

		if err != nil {
			err2, ok := err.(*openwrt.OpenwrtError)
			if ok && err2.Code == 404 {
				_, err3 := handler.CreateObject(clientInfo, new_instance)
				if err3 != nil {
					return false, err3
				}
				changed = true
			} else {
				reqLogger.Error(err, "Failed to get object")
				return false, err
			}
		} else if handler.IsEqual(runtime_instance, new_instance) {
			reqLogger.Info("Equal to the runtime instance, so no update")
		} else {
			_, err := handler.UpdateObject(clientInfo, new_instance)
			if err != nil {
				return false, err
			}
			changed = true
		}
		if changed {
			_, err = handler.Restart(clientInfo)
			if err != nil {
				return changed, err
			}
			cnfChanged = true
		}
	}
	// We say the AddUpdate succeed only when the add/update for all pods succeed
	return cnfChanged, nil
}

func (p *OpenWrtProvider) DeleteObject(handler basehandler.ISdewanHandler, instance runtime.Object) (bool, error) {
	reqLogger := log.WithValues(handler.GetType(), handler.GetName(instance), "cnf", p.Deployment.Name)
	ctx := context.Background()
	podList := &corev1.PodList{}
	err := p.K8sClient.List(ctx, podList, client.MatchingLabels{"sdewanPurpose": p.SdewanPurpose})
	if err != nil {
		return false, err
	}
	cnfChanged := false
	for _, pod := range podList.Items {
		clientInfo := &openwrt.OpenwrtClientInfo{Ip: pod.Status.PodIP, User: "root", Password: ""}
		runtime_instance, err := handler.GetObject(clientInfo, handler.GetName(instance))
		if err != nil {
			err2, ok := err.(*openwrt.OpenwrtError)
			if ok && err2.Code == 404 {
				reqLogger.Info("Runtime instance doesn't exist, so don't have to delete")
				continue
			} else {
				reqLogger.Error(err, "Failed to get object")
				return false, err
			}
		} else if runtime_instance == nil {
			reqLogger.Info("Runtime instance doesn't exist, so don't have to delete")
			continue
		} else {
			err = handler.DeleteObject(clientInfo, handler.GetName(instance))
			if err != nil {
				return false, err
			}
			_, err = handler.Restart(clientInfo)
			if err != nil {
				return false, err
			}
			cnfChanged = true
		}
	}
	// We say the deletioni succeed only when the deletion for all pods succeed
	return cnfChanged, nil
}
