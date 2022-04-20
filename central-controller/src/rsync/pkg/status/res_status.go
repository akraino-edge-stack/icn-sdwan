// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package status

import (
	"encoding/json"

	yaml "github.com/ghodss/yaml"
	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/depend"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotifyserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

var PreInstallHookLabel string = "emco/preinstallHook"

// Update status for the App ready on a cluster and check if app ready on all clusters
func HandleResourcesStatus(acID, app, cluster string, rbData *rb.ResourceBundleState) {

	// Look up the contextId
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(acID)
	if err != nil {
		log.Error("::App context not found::", log.Fields{"acID": acID, "app": app, "cluster": cluster, "err": err})
		return
	}
	// Produce yaml representation of the status
	vjson, err := json.Marshal(rbData.Status)
	if err != nil {
		log.Error("::Error marshalling status information::", log.Fields{"acID": acID, "app": app, "cluster": cluster, "err": err})
		return
	}
	chandle, err := ac.GetClusterHandle(app, cluster)
	if err != nil {
		log.Error("::Error getting cluster handle::", log.Fields{"acID": acID, "app": app, "cluster": cluster, "err": err})
		return
	}
	// Get the handle for the context/app/cluster status object
	handle, _ := ac.GetLevelHandle(chandle, "status")

	// If status handle was not found, then create the status object in the appcontext
	if handle == nil {
		ac.AddLevelValue(chandle, "status", string(vjson))
	} else {
		ac.UpdateStatusValue(handle, string(vjson))
	}

	UpdateAppReadyStatus(acID, app, cluster, rbData)

	// Inform Rsync dependency management of the update
	go depend.ResourcesReady(acID, app, cluster)

	// Send notification to the subscribers
	err = readynotifyserver.SendAppContextNotification(acID)
	if err != nil {
		log.Error("::Error sending ReadyNotify to subscribers::", log.Fields{"acID": acID, "app": app, "cluster": cluster, "err": err})
	}
}

func updateResourcesStatus(acID, app, cluster string, rbData *rb.ResourceBundleState) bool {
	var Ready bool = true
	// Default is ready status
	// In case of Hook resoureces if Pod and Job it is success status
	var statusType types.ResourceStatusType = types.ReadyStatus
	acUtils, err := utils.NewAppContextReference(acID)
	if err != nil {
		return false
	}
	readyChecker := NewReadyChecker(PausedAsReady(true), CheckJobs(true))
	var avail bool = false
	for _, s := range rbData.Status.ServiceStatuses {
		name := s.Name + "+" + "Service"
		b := readyChecker.ServiceReady(&s)
		avail = true
		// If not ready set flag to false
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(types.ReadyStatus), b)
	}
	for _, d := range rbData.Status.DeploymentStatuses {
		avail = true
		name := d.Name + "+" + "Deployment"
		b := readyChecker.DeploymentReady(&d)
		// If not ready set flag to false
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(statusType), b)
	}
	for _, d := range rbData.Status.DaemonSetStatuses {
		avail = true
		name := d.Name + "+" + "Daemon"
		b := readyChecker.DaemonSetReady(&d)
		// If not ready set flag to false
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(statusType), b)
	}
	for _, s := range rbData.Status.StatefulSetStatuses {
		avail = true
		name := s.Name + "+" + "StatefulSet"
		b := readyChecker.StatefulSetReady(&s)
		// If not ready set flag to false
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(types.ReadyStatus), b)
	}
	for _, j := range rbData.Status.JobStatuses {
		name := j.Name + "+" + "Job"
		// Check if the Job is a Hook
		annoMap := j.GetAnnotations()
		_, ok := annoMap["helm.sh/hook"]
		if ok {
			// Hooks are checked for Success Status
			acUtils.SetResourceReadyStatus(app, cluster, name, string(types.SuccessStatus), readyChecker.JobSuccess(&j))
			// No need to consider hook resources for ready status
			continue
		}
		avail = true
		b := readyChecker.JobReady(&j)
		// If not ready set flag to false
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(types.ReadyStatus), b)
	}

	for _, p := range rbData.Status.PodStatuses {
		name := p.Name + "+" + "Pod"

		// Check if the Pod is a Hook
		annoMap := p.GetAnnotations()
		_, ok := annoMap["helm.sh/hook"]
		if ok {
			// Hooks are checked for Success Status
			acUtils.SetResourceReadyStatus(app, cluster, name, string(types.SuccessStatus), readyChecker.PodSuccess(&p))
			// No need to consider hook resources for ready status
			continue
		}
		// Check if Pod is associated with a Job, if so skip that Pod
		// That Pod will never go in ready state
		labels := p.GetLabels()
		_, job := labels["job-name"]
		if job {
			acUtils.SetResourceReadyStatus(app, cluster, name, string(types.SuccessStatus), readyChecker.PodSuccess(&p))
			continue
		}
		avail = true
		// If not ready set flag to false
		b := readyChecker.PodReady(&p)
		if !b {
			Ready = false
		}
		acUtils.SetResourceReadyStatus(app, cluster, name, string(statusType), b)
	}
	if !avail {
		return false
	}

	return Ready

}
func UpdateAppReadyStatus(acID, app string, cluster string, rbData *rb.ResourceBundleState) bool {

	// Check if hook label is present
	labels := rbData.GetLabels()
	// Status tracking is also used for per install hooks
	// At the time of preinstall hook installation cluster
	// ready status should not be updated
	_, hookCR := labels[PreInstallHookLabel]
	acUtils, err := utils.NewAppContextReference(acID)
	if err != nil {
		return false
	}
	if hookCR {
		// If hookCR label, no need to update the ready status
		// Main resources not installed yet
		return updateResourcesStatus(acID, app, cluster, rbData)
	}
	//  Update AppContext to flase
	// If the application is not ready stop processing
	if !updateResourcesStatus(acID, app, cluster, rbData) {
		acUtils.SetClusterResourcesReady(app, cluster, false)
		return false
	}
	// If Application is ready on the cluster, Update AppContext
	acUtils.SetClusterResourcesReady(app, cluster, true)
	log.Info(" UpdateAppReadyStatus:: App is ready on cluster", log.Fields{"acID": acID, "app": app, "cluster": cluster})
	return true
}

// GetStatusCR returns a status monitoring customer resource
func GetStatusCR(label string, extraLabel string) ([]byte, error) {

	var statusCr rb.ResourceBundleState

	statusCr.TypeMeta.APIVersion = "k8splugin.io/v1alpha1"
	statusCr.TypeMeta.Kind = "ResourceBundleState"
	statusCr.SetName(label)

	labels := make(map[string]string)
	labels["emco/deployment-id"] = label
	if len(extraLabel) > 0 {
		labels[extraLabel] = "true"
	}
	statusCr.SetLabels(labels)

	labelSelector, err := metav1.ParseToLabelSelector("emco/deployment-id = " + label)
	if err != nil {
		return nil, err
	}
	statusCr.Spec.Selector = labelSelector

	// Marshaling to json then convert to yaml works better than marshaling to yaml
	// The 'apiVersion' attribute was marshaling to 'apiversion'
	j, err := json.Marshal(&statusCr)
	if err != nil {
		return nil, err
	}
	y, err := yaml.JSONToYAML(j)
	if err != nil {
		return nil, err
	}

	return y, nil
}

//TagResource with label
func TagResource(res []byte, label string) ([]byte, error) {

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAMLData(string(res), unstruct)
	if err != nil {
		return nil, err
	}
	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	//labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	labels["emco/deployment-id"] = label
	unstruct.SetLabels(labels)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	//connector.TagPodsIfPresent(unstruct, client.GetInstanceID())
	TagPodsIfPresent(unstruct, label)
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return b, nil
}

// TagPodsIfPresent finds the TemplateSpec from any workload
// object that contains it and changes the spec to include the tag label
func TagPodsIfPresent(unstruct *unstructured.Unstructured, tag string) {
	_, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template")
	if err != nil || !found {
		return
	}
	// extract spec template labels
	labels, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template", "metadata", "labels")
	if err != nil {
		log.Error("TagPodsIfPresent: Error reading the NestMap for template", log.Fields{"unstruct": unstruct, "err": err})
		return
	}
	if labels == nil || !found {
		labels = make(map[string]interface{})
	}
	labels["emco/deployment-id"] = tag
	if err := unstructured.SetNestedMap(unstruct.Object, labels, "spec", "template", "metadata", "labels"); err != nil {
		log.Error("Error tagging template with emco label", log.Fields{"err": err})
	}
}
