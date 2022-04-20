/*
Copyright The Helm Authors.

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

package status

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
)

// ReadyCheckerOption is a function that configures a ReadyChecker.
type ReadyCheckerOption func(*ReadyChecker)

// PausedAsReady returns a ReadyCheckerOption that configures a ReadyChecker
// to consider paused resources to be ready. For example a Deployment
// with spec.paused equal to true would be considered ready.
func PausedAsReady(pausedAsReady bool) ReadyCheckerOption {
	return func(c *ReadyChecker) {
		c.pausedAsReady = pausedAsReady
	}
}

// CheckJobs returns a ReadyCheckerOption that configures a ReadyChecker
// to consider readiness of Job resources.
func CheckJobs(checkJobs bool) ReadyCheckerOption {
	return func(c *ReadyChecker) {
		c.checkJobs = checkJobs
	}
}

// NewReadyChecker creates a new checker. Passed ReadyCheckerOptions can
// be used to override defaults.
func NewReadyChecker(opts ...ReadyCheckerOption) ReadyChecker {
	c := ReadyChecker{}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// ReadyChecker is a type that can check core Kubernetes types for readiness.
type ReadyChecker struct {
	checkJobs     bool
	pausedAsReady bool
}

// PodReady returns true if a pod is ready; false otherwise.
func (c *ReadyChecker) PodReady(pod *corev1.Pod) bool {
	for _, cn := range pod.Status.Conditions {
		if cn.Type == corev1.PodReady && cn.Status == corev1.ConditionTrue {
			return true
		}
	}
	log.Printf("Pod is not ready: %s/%s", pod.GetNamespace(), pod.GetName())
	return false
}

func (c *ReadyChecker) JobReady(job *batchv1.Job) bool {
	if c.checkJobs {
		if job.Status.Failed > *job.Spec.BackoffLimit {
			log.Printf("Job is failed: %s/%s", job.GetNamespace(), job.GetName())
			return false
		}
		if job.Status.Succeeded < *job.Spec.Completions {
			log.Printf("Job is not completed: %s/%s", job.GetNamespace(), job.GetName())
			return false
		}
		return true
	}
	return false
}

func (c *ReadyChecker) ServiceReady(s *corev1.Service) bool {
	// ExternalName Services are external to cluster so helm shouldn't be checking to see if they're 'ready' (i.e. have an IP Set)
	if s.Spec.Type == corev1.ServiceTypeExternalName {
		return true
	}

	// Ensure that the service cluster IP is not empty
	if s.Spec.ClusterIP == "" {
		log.Printf("Service does not have cluster IP address: %s/%s", s.GetNamespace(), s.GetName())
		return false
	}

	// This checks if the service has a LoadBalancer and that balancer has an Ingress defined
	if s.Spec.Type == corev1.ServiceTypeLoadBalancer {
		// do not wait when at least 1 external IP is set
		if len(s.Spec.ExternalIPs) > 0 {
			log.Printf("Service %s/%s has external IP addresses (%v), marking as ready", s.GetNamespace(), s.GetName(), s.Spec.ExternalIPs)
			return true
		}

		if s.Status.LoadBalancer.Ingress == nil {
			log.Printf("Service does not have load balancer ingress IP address: %s/%s", s.GetNamespace(), s.GetName())
			return false
		}
	}

	return true
}

func (c *ReadyChecker) VolumeReady(v *corev1.PersistentVolumeClaim) bool {
	if v.Status.Phase != corev1.ClaimBound {
		log.Printf("PersistentVolumeClaim is not bound: %s/%s", v.GetNamespace(), v.GetName())
		return false
	}
	return true
}

func (c *ReadyChecker) DeploymentReady(dep *appsv1.Deployment) bool {
	// If paused deployment will never be ready
	if dep.Spec.Paused {
		return c.pausedAsReady
	}
	expectedReady := *dep.Spec.Replicas - MaxUnavailable(*dep)
	if !(dep.Status.ReadyReplicas >= expectedReady) {
		log.Printf("Deployment is not ready: %s/%s. %d out of %d expected pods are ready", dep.Namespace, dep.Name, dep.Status.ReadyReplicas, expectedReady)
		return false
	}
	return true
}

func (c *ReadyChecker) DaemonSetReady(ds *appsv1.DaemonSet) bool {
	// If the update strategy is not a rolling update, there will be nothing to wait for
	if ds.Spec.UpdateStrategy.Type != appsv1.RollingUpdateDaemonSetStrategyType {
		return true
	}

	// Make sure all the updated pods have been scheduled
	if ds.Status.UpdatedNumberScheduled != ds.Status.DesiredNumberScheduled {
		log.Printf("DaemonSet is not ready: %s/%s. %d out of %d expected pods have been scheduled", ds.Namespace, ds.Name, ds.Status.UpdatedNumberScheduled, ds.Status.DesiredNumberScheduled)
		return false
	}
	maxUnavailable, err := intstr.GetValueFromIntOrPercent(ds.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable, int(ds.Status.DesiredNumberScheduled), true)
	if err != nil {
		// If for some reason the value is invalid, set max unavailable to the
		// number of desired replicas. This is the same behavior as the
		// `MaxUnavailable` function in deploymentutil
		maxUnavailable = int(ds.Status.DesiredNumberScheduled)
	}

	expectedReady := int(ds.Status.DesiredNumberScheduled) - maxUnavailable
	if !(int(ds.Status.NumberReady) >= expectedReady) {
		log.Printf("DaemonSet is not ready: %s/%s. %d out of %d expected pods are ready", ds.Namespace, ds.Name, ds.Status.NumberReady, expectedReady)
		return false
	}
	return true
}

func (c *ReadyChecker) StatefulSetReady(sts *appsv1.StatefulSet) bool {
	// If the update strategy is not a rolling update, there will be nothing to wait for
	if sts.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		return true
	}

	// Dereference all the pointers because StatefulSets like them
	var partition int
	// 1 is the default for replicas if not set
	var replicas = 1
	// For some reason, even if the update strategy is a rolling update, the
	// actual rollingUpdate field can be nil. If it is, we can safely assume
	// there is no partition value
	if sts.Spec.UpdateStrategy.RollingUpdate != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
		partition = int(*sts.Spec.UpdateStrategy.RollingUpdate.Partition)
	}
	if sts.Spec.Replicas != nil {
		replicas = int(*sts.Spec.Replicas)
	}

	// Because an update strategy can use partitioning, we need to calculate the
	// number of updated replicas we should have. For example, if the replicas
	// is set to 3 and the partition is 2, we'd expect only one pod to be
	// updated
	expectedReplicas := replicas - partition

	// Make sure all the updated pods have been scheduled
	if int(sts.Status.UpdatedReplicas) != expectedReplicas {
		log.Printf("StatefulSet is not ready: %s/%s. %d out of %d expected pods have been scheduled", sts.Namespace, sts.Name, sts.Status.UpdatedReplicas, expectedReplicas)
		return false
	}

	if int(sts.Status.ReadyReplicas) != replicas {
		log.Printf("StatefulSet is not ready: %s/%s. %d out of %d expected pods are ready", sts.Namespace, sts.Name, sts.Status.ReadyReplicas, replicas)
		return false
	}
	return true
}

// These methods are mainly for hook implementations.
//
// For most kinds, the checks to see if the resource is marked as Added or Modified
// by the Kubernetes event stream is enough. For some kinds, more is required:
//
// - Jobs: A job is marked "Ready" when it has successfully completed. This is
//   ascertained by watching the Status fields in a job's output.
// - Pods: A pod is marked "Ready" when it has successfully completed. This is
//   ascertained by watching the status.phase field in a pod's output.
//
func (c *ReadyChecker) PodSuccess(pod *corev1.Pod) bool {

	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		logutils.Info("Pod succeeded::", logutils.Fields{"Pod Name": pod.Name})
		return true
	case corev1.PodFailed:
		return true
	case corev1.PodPending:
		logutils.Info("Pod running::", logutils.Fields{"Pod Name": pod.Name})
	case corev1.PodRunning:
		logutils.Info("Pod is Running::", logutils.Fields{"Pod Name": pod.Name})
	}

	return false
}

func (c *ReadyChecker) JobSuccess(job *batchv1.Job) bool {

	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == "True" {
			return true
		} else if c.Type == batchv1.JobFailed && c.Status == "True" {
			return true
		}
	}
	logutils.Info("Job Status:", logutils.Fields{"Jobs active": job.Status.Active, "Jobs failed": job.Status.Failed, "Jobs Succeded": job.Status.Succeeded})
	return false
}
