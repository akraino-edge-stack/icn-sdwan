// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"context"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/resourcestatus"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	"strings"
	"time"
)

type resProvd struct {
	app     string
	cluster string
	cl      ClientProvider
	context Context
}

// Hook Kinds that require wait
var HelmHookRelevantKindSet map[string]bool = map[string]bool{"Job": true, "Pod": true, "Deployment": true, "DaemonSet": true, "StatefulSet": true}

// handleResources - performs the operation on the cluster
// If wait is true wait for resources to be ready/success before next resource
func (r *resProvd) handleResources(ctx context.Context, op RsyncOperation, resources []string) (int, error) {

	var handledRes int
	var ref interface{}
	var breakonError bool

	if len(resources) <= 0 {
		return handledRes, nil
	}
	// Keep retrying for reachability
	for {
		// Wait for cluster to be reachable
		err := r.waitForClusterReady(ctx)
		if err != nil {
			return handledRes, err
		}
		reachable := true
		handledRes = 0
		// Handle all resources in order
		for _, res := range resources {
			ref, breakonError, err = r.handleResource(ctx, op, res, ref)
			if err != nil {
				log.Error("Error in resource", log.Fields{"error": err, "cluster": r.cluster, "resource": res})
				// If failure is due to reachability issues start retrying
				if err1 := r.cl.IsReachable(); err1 != nil {
					reachable = false
					break
				}
				if breakonError {
					return handledRes, err
				}
			}
			handledRes++
		}
		// Check if the break from loop due to reachabilty issues
		if reachable {
			// Not reachability issue, commit resources to the cluster
			r.cl.Commit(ctx, ref)
			return handledRes, nil
		}
	}
}

func (r *resProvd) handleResourcesWithWait(ctx context.Context, op RsyncOperation, resources []string) (int, error) {
	var handledRes int
	for _, res := range resources {
		// Apply one resource and then wait for it to be succeded before going to next resource
		if _, err := r.handleResources(ctx, op, []string{res}); err == nil {
			// Resouce successfully applied
			handledRes++
			// Check if this resource needs waiting based on Helm comment:
			// https://github.com/helm/helm/blob/67f92d63faef0f9acdd0a95680dbc03dc7f86ed6/pkg/kube/client.go#L552
			// For things like a secret or a config map, this is the best indicator
			// we get. We care mostly about jobs, where what we want to see is
			// the status go into a good state.
			result := strings.SplitN(res, "+", 2)
			if len(result) != 2 {
				log.Error("Invalid resource name format::", log.Fields{"res": res})
				return handledRes, err
			}
			if ok := HelmHookRelevantKindSet[result[1]]; !ok {
				continue
			}
			if err := r.context.dm.WaitResourceDependency(ctx, r.app, r.cluster, res); err != nil {
				log.Error("Dependency error", log.Fields{"error": err, "cluster": r.cluster, "resource": res})
				return handledRes, err
			}
		} else {
			return handledRes, err
		}
	}
	return handledRes, nil
}

func (r *resProvd) waitForClusterReady(ctx context.Context) error {

	// Check if reachable
	if err := r.cl.IsReachable(); err == nil {
		r.context.acRef.SetClusterAvailableStatus(r.app, r.cluster, appcontext.ClusterReadyStatusEnum.Available)
		return nil
	}
	r.context.acRef.SetClusterAvailableStatus(r.app, r.cluster, appcontext.ClusterReadyStatusEnum.Retrying)
	timedOut := false
	retryCnt := 0
	forceDone := false
Loop:
	for {
		select {
		// Wait for wait time before checking cluster ready
		case <-time.After(time.Duration(r.context.waitTime) * time.Second):
			// Context is canceled
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// If cluster is reachable then done
			if err := r.cl.IsReachable(); err == nil {
				r.context.acRef.SetClusterAvailableStatus(r.app, r.cluster, appcontext.ClusterReadyStatusEnum.Available)
				return nil
			}
			log.Info("Cluster is not reachable - keep trying::", log.Fields{"cluster": r.cluster, "retry count": retryCnt})
			retryCnt++
			if r.context.maxRetry >= 0 && retryCnt > r.context.maxRetry {
				timedOut = true
				break Loop
			}
		// Check if the context is canceled
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if timedOut {
		return pkgerrors.Errorf("Retries exceeded max: " + r.cluster)
	}
	if forceDone {
		return pkgerrors.Errorf("Termination of rsync cluster retry: " + r.cluster)
	}
	return nil
}

func (r *resProvd) handleResource(ctx context.Context, op RsyncOperation, res string, ref interface{}) (interface{}, bool, error) {

	var q interface{}
	var err error
	switch op {
	case OpApply:
		// Get resource dependency here
		q, err = r.instantiateResource(res, ref)
		if err != nil {
			// return true for breakon error
			return q, true, err
		}
	case OpCreate:
		// Get resource dependency here
		q, err = r.createResource(res, ref)
		if err != nil {
			// return true for breakon error
			return q, true, err
		}
	case OpDelete:
		q, err = r.terminateResource(res, ref)
		if err != nil {
			// return false for breakon error
			return q, false, err
		}
	case OpRead:
		err = r.readResource(res)
		if err != nil {
			// return false for breakon error
			return nil, false, err
		}
	}
	// return false for breakon error
	return q, false, nil
}

func (r *resProvd) instantiateResource(name string, ref interface{}) (interface{}, error) {
	var q interface{}
	res, _, err := r.context.acRef.GetRes(name, r.app, r.cluster)
	if err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		return nil, err
	}
	label := r.context.statusAcID + "-" + r.app
	b, err := r.cl.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": name})
		return nil, err
	}
	if q, err = r.cl.Apply(name, ref, b); err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to apply res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	return q, nil
}

func (r *resProvd) createResource(name string, ref interface{}) (interface{}, error) {
	var q interface{}

	res, _, err := r.context.acRef.GetRes(name, r.app, r.cluster)
	if err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		return nil, err
	}
	label := r.context.statusAcID + "-" + r.app
	b, err := r.cl.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": name})
		return nil, err
	}
	if q, err = r.cl.Create(name, ref, b); err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to create res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	return q, nil
}

func (r *resProvd) terminateResource(name string, ref interface{}) (interface{}, error) {
	var q interface{}

	res, sh, err := r.context.acRef.GetRes(name, r.app, r.cluster)
	if err != nil {
		if sh != nil {
			r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		}
		return nil, err
	}
	if q, err = r.cl.Delete(name, ref, res); err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to delete res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Deleted})
	return q, nil
}

func (r *resProvd) readResource(name string) error {

	res, _, err := r.context.acRef.GetRes(name, r.app, r.cluster)
	if err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		return err
	}
	// Get the resource from the cluster
	b, err := r.cl.Get(name, res)
	if err != nil {
		r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Failed})
		log.Error("Failed to read res", log.Fields{"error": err, "resource": name})
		return err
	}
	// Store result back in AppContext
	r.context.acRef.PutRes(name, r.app, r.cluster, b)
	r.updateResourceStatus(name, resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
	return nil
}

func (r *resProvd) addStatusTracker(extraLabel string) error {
	// Get label and create CR
	label := r.context.statusAcID + "-" + r.app
	b, err := status.GetStatusCR(label, extraLabel)
	if err != nil {
		log.Error("Failed to get status CR for installing", log.Fields{"error": err, "label": label})
		return err
	}
	if err = r.cl.ApplyStatusCR(label, b); err != nil {
		log.Error("Failed to apply status tracker", log.Fields{"error": err, "cluster": r.cluster, "app": r.app, "label": label})
		return err
	}
	return nil
}

func (r *resProvd) deleteStatusTracker(extraLabel string) error {
	// Get label and create CR
	label := r.context.statusAcID + "-" + r.app
	b, err := status.GetStatusCR(label, extraLabel)
	if err != nil {
		log.Error("Failed to get status CR for deleting", log.Fields{"error": err, "label": label})
		return err
	}
	if err = r.cl.DeleteStatusCR(label, b); err != nil {
		log.Error("Failed to delete res", log.Fields{"error": err, "app": r.app, "label": label})
		return err
	}
	return nil
}

func (r *resProvd) updateResourceStatus(name string, status interface{}) {
	// Use utils with status appContext
	_ = r.context.scRef.AddResourceStatus(name, r.app, r.cluster, status, r.context.acID)
	// Treating status errors as non fatal
}
