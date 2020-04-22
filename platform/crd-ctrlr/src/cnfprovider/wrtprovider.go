package cnfprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"reflect"
	sdewanv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

var log = logf.Log.WithName("wrt_provider")

type WrtProvider struct {
	Namespace     string
	SdewanPurpose string
	Deployment    extensionsv1beta1.Deployment
	K8sClient     client.Client
}

func NewWrt(namespace string, sdewanPurpose string, k8sClient client.Client) (*WrtProvider, error) {
	reqLogger := log.WithValues("namespace", namespace, "sdewanPurpose", sdewanPurpose)
	ctx := context.Background()
	deployments := &extensionsv1beta1.DeploymentList{}
	err := k8sClient.List(ctx, deployments, client.MatchingLabels{"sdewanPurpose": sdewanPurpose})
	if err != nil {
		reqLogger.Error(err, "Failed to get cnf deployment")
		return nil, client.IgnoreNotFound(err)
	}
	if len(deployments.Items) != 1 {
		reqLogger.Error(nil, "More than one deployment exists")
		return nil, errors.New("More than one deployment exists")
	}

	return &WrtProvider{namespace, sdewanPurpose, deployments.Items[0], k8sClient}, nil
}

func (p *WrtProvider) net2iface(net string) (string, error) {
	type Iface struct {
		DefaultGateway bool `json:"defaultGateway,string"`
		Interface      string
		Name           string
	}
	type NfnNet struct {
		Type      string
		Interface []Iface
	}
	ann := p.Deployment.Spec.Template.Annotations
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

func (p *WrtProvider) convertCrd(mwan3Policy *sdewanv1alpha1.Mwan3Policy) (*openwrt.SdewanPolicy, error) {
	members := make([]openwrt.SdewanMember, len(mwan3Policy.Spec.Members))
	for i, membercr := range mwan3Policy.Spec.Members {
		iface, err := p.net2iface(membercr.Network)
		if err != nil {
			return nil, err
		}
		members[i] = openwrt.SdewanMember{
			Interface: iface,
			Metric:    strconv.Itoa(membercr.Metric),
			Weight:    strconv.Itoa(membercr.Weight),
		}
	}
	return &openwrt.SdewanPolicy{Name: mwan3Policy.Name, Members: members}, nil

}

func (p *WrtProvider) AddUpdateMwan3Policy(mwan3Policy *sdewanv1alpha1.Mwan3Policy) (bool, error) {
	reqLogger := log.WithValues("Mwan3Policy", mwan3Policy.Name, "cnf", p.Deployment.Name)
	ctx := context.Background()
	podList := &corev1.PodList{}
	err := p.K8sClient.List(ctx, podList, client.MatchingLabels{"sdewanPurpose": p.SdewanPurpose})
	if err != nil {
		reqLogger.Error(err, "Failed to get cnf pod list")
		return false, err
	}
	policy, err := p.convertCrd(mwan3Policy)
	if err != nil {
		reqLogger.Error(err, "Failed to convert mwan3Policy CR")
		return false, err
	}
	cnfChanged := false
	for _, pod := range podList.Items {
		openwrtClient := openwrt.NewOpenwrtClient(pod.Status.PodIP, "root", "")
		mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
		service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
		runtimePolicy, _ := mwan3.GetPolicy(policy.Name)
		changed := false
		if runtimePolicy == nil {
			_, err := mwan3.CreatePolicy(*policy)
			if err != nil {
				reqLogger.Error(err, "Failed to create policy")
				return false, err
			}
			changed = true
		} else if reflect.DeepEqual(*runtimePolicy, *policy) {
			reqLogger.Info("Equal to the runtime policy, so no update")
		} else {
			_, err := mwan3.UpdatePolicy(*policy)
			if err != nil {
				reqLogger.Error(err, "Failed to update policy")
				return false, err
			}
			changed = true
		}
		if changed {
			_, err = service.ExecuteService("mwan3", "restart")
			if err != nil {
				reqLogger.Error(err, "Failed to restart mwan3 service")
				return changed, err
			}
			cnfChanged = true
		}
	}
	// We say the AddUpdate succeed only when the add/update for all pods succeed
	return cnfChanged, nil
}

func (p *WrtProvider) DeleteMwan3Policy(mwan3Policy *sdewanv1alpha1.Mwan3Policy) (bool, error) {
	reqLogger := log.WithValues("Mwan3Policy", mwan3Policy.Name, "cnf", p.Deployment.Name)
	ctx := context.Background()
	podList := &corev1.PodList{}
	err := p.K8sClient.List(ctx, podList, client.MatchingLabels{"sdewanPurpose": p.SdewanPurpose})
	if err != nil {
		reqLogger.Error(err, "Failed to get pod list")
		return false, err
	}
	cnfChanged := false
	for _, pod := range podList.Items {
		openwrtClient := openwrt.NewOpenwrtClient(pod.Status.PodIP, "root", "")
		mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
		service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
		runtimePolicy, _ := mwan3.GetPolicy(mwan3Policy.Name)
		if runtimePolicy == nil {
			reqLogger.Info("Runtime policy doesn't exist, so don't have to delete")
		} else {
			err = mwan3.DeletePolicy(mwan3Policy.Name)
			if err != nil {
				reqLogger.Error(err, "Failed to delete policy")
				return false, err
			}
			_, err = service.ExecuteService("mwan3", "restart")
			if err != nil {
				reqLogger.Error(err, "Failed to restart mwan3 service")
				return false, err
			}
			cnfChanged = true
		}
	}
	// We say the deletioni succeed only when the deletion for all pods succeed
	return cnfChanged, nil
}
