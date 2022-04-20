// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8s

import (
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	v1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	clientset "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/client/clientset/versioned"
	informers "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/client/informers/externalversions"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type channelManager struct {
	channels map[string]chan struct{}
	sync.Mutex
}

var channelData channelManager

const monitorLabel = "emco/deployment-id"

// HandleStatusUpdate for an application in a cluster
func HandleStatusUpdate(clusterId string, id string, v *v1alpha1.ResourceBundleState) {
	log.Info("K8s HandleStatusUpdate", log.Fields{"id": id, "cluster": clusterId})
	// Get the contextId from the label (id)
	result := strings.SplitN(id, "-", 2)
	log.Info("::HandleStatusUpdate::", log.Fields{"id": id, "cluster": clusterId})
	if result[0] == "" {
		log.Error("::label is missing an appcontext identifier::", log.Fields{"id": id, "cluster": clusterId})
		return
	}
	if len(result) != 2 {
		log.Error("::invalid label format::", log.Fields{"id": id, "cluster": clusterId})
		return
	}
	// Get the app from the label (id)
	if result[1] == "" {
		log.Error("::label is missing an app identifier::", log.Fields{"id": id, "cluster": clusterId})
		return
	}
	log.Info("K8s HandleStatusUpdate", log.Fields{"id": id, "cluster": clusterId, "app": result[1]})
	// Notify Resource tracking
	status.HandleResourcesStatus(result[0], result[1], clusterId, v)
}

// StartClusterWatcher watches for CR
// configBytes - Kubectl file data
func (c *K8sProvider) StartClusterWatcher() error {

	// a cluster watcher always watches the cluster as a whole, so rsync's CloudConfig level
	// is 0 and namespace doesn't need to be specified because the result is non-ambiguous
	log.Info("Starting cluster watcher with L0 cloudconfig", log.Fields{})

	//key := provider + "+" + name
	// Get the lock
	channelData.Lock()
	defer channelData.Unlock()
	// For first time
	if channelData.channels == nil {
		channelData.channels = make(map[string]chan struct{})
	}
	_, ok := channelData.channels[c.cluster]
	if !ok {
		// Create Channel
		channelData.channels[c.cluster] = make(chan struct{})
		// Read config
		configBytes, err := utils.GetKubeConfig(c.cluster, "0", "")
		if err != nil {
			return err
		}
		// Create config
		config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
		if err != nil {
			log.Error("RESTConfigFromKubeConfig error:", log.Fields{"err": err})
			return pkgerrors.Wrap(err, "RESTConfigFromKubeConfig error")
		}
		k8sClient, err := clientset.NewForConfig(config)
		if err != nil {
			return pkgerrors.Wrap(err, "Clientset NewForConfig error")
		}
		// Create Informer
		mInformerFactory := informers.NewSharedInformerFactory(k8sClient, 0)
		mInformer := mInformerFactory.K8splugin().V1alpha1().ResourceBundleStates().Informer()
		go scheduleStatus(c.cluster, channelData.channels[c.cluster], mInformer)
	}
	return nil
}

// StopClusterWatcher stop watching a cluster
func StopClusterWatcher(clusterId string) {
	//key := provider + "+" + name
	if channelData.channels != nil {
		c, ok := channelData.channels[clusterId]
		if ok {
			close(c)
		}
	}
}

// CloseAllClusterWatchers close all channels
func CloseAllClusterWatchers() {
	if channelData.channels == nil {
		return
	}
	// Close all Channels to stop all watchers
	for _, e := range channelData.channels {
		close(e)
	}
}

// Per Cluster Go routine to watch CR
func scheduleStatus(clusterId string, c <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(clusterId, l, v)
				}
			}
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(clusterId, l, v)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// Ignore it
		},
	}
	s.AddEventHandler(handlers)
	s.Run(c)
}

// ApplyStatusCR applies status CR
func (p *K8sProvider) ApplyStatusCR(name string, content []byte) error {
	if err := p.client.Apply(content); err != nil {
		log.Error("Failed to apply Status CR", log.Fields{
			"error": err,
		})
		return err
	}
	return nil

}

// DeleteStatusCR deletes status CR
func (p *K8sProvider) DeleteStatusCR(name string, content []byte) error {
	if err := p.client.Delete(content); err != nil {
		log.Error("Failed to delete Status CR", log.Fields{
			"error": err,
		})
		return err
	}
	return nil
}
