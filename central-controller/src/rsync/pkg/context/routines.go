// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"bytes"
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/depend"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	"golang.org/x/sync/errgroup"
)

// Check status of AppContext against the event to see if it is valid
func (c *Context) checkStateChange(e RsyncEvent) (StateChange, bool, error) {
	var supported bool = false
	var err error
	var dState, cState appcontext.AppContextStatus
	var event StateChange

	event, ok := StateChanges[e]
	if !ok {
		return StateChange{}, false, pkgerrors.Errorf("Invalid Event %s:", e)
	}
	// Check Stop flag return error no processing desired
	sFlag, err := c.acRef.GetAppContextFlag(StopFlagKey)
	if err != nil {
		return event, false, pkgerrors.Errorf("AppContext Error: %s:", err)
	}
	if sFlag {
		return event, false, pkgerrors.Errorf("Stop flag set for context: %s", c.acID)
	}
	// Check PendingTerminate Flag
	tFlag, err := c.acRef.GetAppContextFlag(PendingTerminateFlagKey)
	if err != nil {
		return event, false, pkgerrors.Errorf("AppContext Error: %s:", err)
	}

	if tFlag && e != TerminateEvent {
		return event, false, pkgerrors.Errorf("Terminate Flag is set, Ignoring event: %s:", e)
	}
	// Update the desired state of the AppContext based on this event
	state, err := c.acRef.GetAppContextStatus(CurrentStateKey)
	if err != nil {
		return event, false, err
	}
	for _, s := range event.SState {
		if s == state.Status {
			supported = true
			break
		}
	}
	if !supported {
		//Exception to state machine, if event is terminate and the current state is already
		// terminated, don't change the status to TerminateFailed
		if e == TerminateEvent && state.Status == appcontext.AppContextStatusEnum.Terminated {
			return event, false, pkgerrors.Errorf("Invalid Source state %s for the Event %s:", state, e)
		}
		return event, true, pkgerrors.Errorf("Invalid Source state %s for the Event %s:", state, e)
	} else {
		dState.Status = event.DState
		cState.Status = event.CState
	}
	// Event is supported. Update Desired state and current state
	err = c.acRef.UpdateAppContextStatus(DesiredStateKey, dState)
	if err != nil {
		return event, false, err
	}
	err = c.acRef.UpdateAppContextStatus(CurrentStateKey, cState)
	if err != nil {
		return event, false, err
	}
	err = c.acRef.UpdateAppContextStatus(StatusKey, cState)
	if err != nil {
		return event, false, err
	}
	return event, true, nil
}

// UpdateQStatus updates status of an element in the queue
func (c *Context) UpdateQStatus(index int, status string) error {
	qUtils := &AppContextQueueUtils{ac: c.acRef.GetAppContextHandle()}
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if err := qUtils.UpdateStatus(index, status); err != nil {
		return err
	}
	return nil
}

// If terminate event recieved set flag to ignore other events
func (c *Context) terminateContextRoutine() {
	// Set Terminate Flag to Pending
	if err := c.acRef.UpdateAppContextFlag(PendingTerminateFlagKey, true); err != nil {
		return
	}
	// Make all waiting goroutines to stop waiting
	if c.cancel != nil {
		c.cancel()
	}
}

// Start Main Thread for handling
func (c *Context) startMainThread(a interface{}, con Connector) error {
	acID := fmt.Sprintf("%v", a)

	ref, err := utils.NewAppContextReference(acID)
	if err != nil {
		return err
	}
	// Read AppContext into CompositeApp structure
	c.ca, err = ReadAppContext(a)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	c.acID = acID
	c.con = con
	c.acRef = ref
	// Wait for 2 secs
	c.waitTime = 2
	c.maxRetry = getMaxRetries()
	// Check flags in AppContext to create if they don't exist and add default values
	_, err = c.acRef.GetAppContextStatus(CurrentStateKey)
	// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
	if err != nil {
		as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
		if err := c.acRef.UpdateAppContextStatus(CurrentStateKey, as); err != nil {
			return err
		}
	}
	_, err = c.acRef.GetAppContextFlag(StopFlagKey)
	// Assume doesn't exist and add
	if err != nil {
		if err := c.acRef.UpdateAppContextFlag(StopFlagKey, false); err != nil {
			return err
		}
	}
	_, err = c.acRef.GetAppContextFlag(PendingTerminateFlagKey)
	// Assume doesn't exist and add
	if err != nil {
		if err := c.acRef.UpdateAppContextFlag(PendingTerminateFlagKey, false); err != nil {
			return err
		}
	}
	_, err = c.acRef.GetAppContextStatus(StatusKey)
	// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
	if err != nil {
		as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
		if err := c.acRef.UpdateAppContextStatus(StatusKey, as); err != nil {
			return err
		}
	}
	// Read the statusAcID to use with status
	c.statusAcID, err = c.acRef.GetStatusAppContext(StatusAppContextIDKey)
	if err != nil {
		// Use appcontext as status appcontext also
		c.statusAcID = c.acID
		c.scRef = c.acRef
	} else {
		scRef, err := utils.NewAppContextReference(c.statusAcID)
		if err != nil {
			return err
		}
		c.scRef = scRef
	}
	// Intialize dependency management
	c.dm = depend.NewDependManager(c.acID)
	// Start Routine to handle AppContext
	go c.appContextRoutine()
	return nil
}

// Handle AppContext
func (c *Context) appContextRoutine() {
	var lctx context.Context
	var l context.Context
	var lGroup *errgroup.Group
	var lDone context.CancelFunc
	var op RsyncOperation

	qUtils := &AppContextQueueUtils{ac: c.acRef.GetAppContextHandle()}
	// Create context for the running threads
	ctx, done := context.WithCancel(context.Background())
	gGroup, gctx := errgroup.WithContext(ctx)
	// Stop all running goroutines
	defer done()
	// Start thread to watch for external stop flag
	gGroup.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				flag, err := c.acRef.GetAppContextFlag(StopFlagKey)
				if err != nil {
					done()
				} else if flag == true {
					log.Info("Forced stop context", log.Fields{})
					// Forced Stop from outside
					done()
				}
			case <-gctx.Done():
				log.Info("Context done", log.Fields{})
				return gctx.Err()
			}
		}
	})
	// Go over all messages
	for {
		// Get first event to process
		c.Lock.Lock()
		index, ele := qUtils.FindFirstPending()
		if index >= 0 {
			c.Lock.Unlock()
			e := ele.Event
			state, skip, err := c.checkStateChange(e)
			// Event is not valid event for the current state of AppContext
			if err != nil {
				log.Error("State Change Error", log.Fields{"error": err})
				if err := c.UpdateQStatus(index, "Skip"); err != nil {
					break
				}
				if !skip {
					// Update status with error
					err = c.acRef.UpdateAppContextStatus(StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
					if err != nil {
						break
					}
					// Update Current status with error
					err = c.acRef.UpdateAppContextStatus(CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
					if err != nil {
						break
					}
				}
				// Continue to process more events
				continue
			}
			// Create a derived context
			l, lDone = context.WithCancel(ctx)
			lGroup, lctx = errgroup.WithContext(l)
			c.Lock.Lock()
			c.cancel = lDone
			c.Lock.Unlock()
			switch e {
			case InstantiateEvent:
				op = OpApply
			case TerminateEvent:
				op = OpDelete
			case ReadEvent:
				op = OpRead
			case UpdateEvent:
				// In Instantiate Phase find out resources that need to be modified and
				// set skip to be true for those that match
				// This is done to avoid applying resources that have no differences
				if err := c.updateModifyPhase(ele); err != nil {
					break
				}
				op = OpApply
				// Enqueue Modify Phase for the AppContext that is being updated to
				go HandleAppContext(ele.UCID, c.acID, UpdateDeleteEvent, c.con)
			case UpdateDeleteEvent:
				// Update AppContext to decide what needs update
				if err := c.updateDeletePhase(ele); err != nil {
					break
				}
				op = OpDelete
			case AddChildContextEvent:
				log.Error("Not Implemented", log.Fields{"event": e})
				if err := c.UpdateQStatus(index, "Skip"); err != nil {
					break
				}
				continue
			default:
				log.Error("Unknown event", log.Fields{"event": e})
				if err := c.UpdateQStatus(index, "Skip"); err != nil {
					break
				}
				continue
			}
			lGroup.Go(func() error {
				return c.run(lctx, lGroup, op, e)
			})
			// Wait for all subtasks to complete
			log.Info("Wait for all subtasks to complete", log.Fields{})
			if err := lGroup.Wait(); err != nil {
				log.Error("Failed run", log.Fields{"error": err})
				// Mark the event in Queue
				if err := c.UpdateQStatus(index, "Error"); err != nil {
					break
				}
				// Update failed status
				err = c.acRef.UpdateAppContextStatus(StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				// Update Current status with error
				err = c.acRef.UpdateAppContextStatus(CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				continue
			}
			log.Info("Success all subtasks completed", log.Fields{})
			// Mark the event in Queue
			if err := c.UpdateQStatus(index, "Done"); err != nil {
				break
			}

			// Success - Update Status for the AppContext to match the Desired State
			ds, _ := c.acRef.GetAppContextStatus(DesiredStateKey)
			err = c.acRef.UpdateAppContextStatus(StatusKey, ds)
			err = c.acRef.UpdateAppContextStatus(CurrentStateKey, ds)

		} else {
			// Done Processing all elements in queue
			log.Info("Done Processing - no new messages", log.Fields{"context": c.acID})
			// Set the TerminatePending Flag to false before exiting
			_ = c.acRef.UpdateAppContextFlag(PendingTerminateFlagKey, false)
			// release the active contextIDs
			ok, err := DeleteActiveContextRecord(c.acID)
			if !ok {
				log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
			}
			c.Running = false
			c.Lock.Unlock()
			return
		}
	}
	// Any error in reading/updating appContext is considered
	// fatal and all processing stopped for the AppContext
	// Set running flag to false before exiting
	c.Lock.Lock()
	// release the active contextIDs
	ok, err := DeleteActiveContextRecord(c.acID)
	if !ok {
		log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
	}
	c.Running = false
	c.Lock.Unlock()
}

// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be deleted
func (c *Context) updateDeletePhase(e AppContextQueueElement) error {

	// Read Update AppContext into CompositeApp structure
	uca, err := ReadAppContext(e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that shouldn't be deleted
	for _, app := range c.ca.Apps {
		foundApp := FindApp(uca, app.Name)
		// If app not found that will be deleted (skip false)
		if foundApp {
			// Check if any clusters are deleted
			for _, cluster := range app.Clusters {
				foundCluster := FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					// Check if any resources are deleted
					var resCnt int = 0
					for _, res := range cluster.Resources {
						foundRes := FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// If resource found in both appContext don't delete it
							res.Skip = true
						} else {
							// Resource found to be deleted
							resCnt++
						}
					}
					// No resources marked for deletion, mark this cluster for not deleting
					if resCnt == 0 {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be Modified
func (c *Context) updateModifyPhase(e AppContextQueueElement) error {
	// Read Update from AppContext into CompositeApp structure
	uca, err := ReadAppContext(e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	//acUtils := &utils.AppContextUtils{Ac: c.ac}
	// Load update appcontext also
	uRef, err := utils.NewAppContextReference(e.UCID)
	if err != nil {
		return err
	}
	//updateUtils := &utils.AppContextUtils{Ac: uac}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that match exactly and shouldn't be changed
	for _, app := range c.ca.Apps {
		foundApp := FindApp(uca, app.Name)
		if foundApp {
			// Check if any clusters are modified
			for _, cluster := range app.Clusters {
				foundCluster := FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					diffRes := false
					// Check if any resources are added or modified
					for _, res := range cluster.Resources {
						foundRes := FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// Read the resource from both AppContext and Compare
							cRes, _, err1 := c.acRef.GetRes(res.Name, app.Name, cluster.Name)
							uRes, _, err2 := uRef.GetRes(res.Name, app.Name, cluster.Name)
							if err1 != nil || err2 != nil {
								log.Error("Fatal Error: reading resources", log.Fields{"err1": err1, "err2": err2})
								return err1
							}
							if bytes.Equal(cRes, uRes) {
								res.Skip = true
							} else {
								log.Info("Update Resource Diff found::", log.Fields{"resource": res.Name, "cluster": cluster})
								diffRes = true
							}
						} else {
							// Found a new resource that is added to the cluster
							diffRes = true
						}
					}
					// If no resources diff, skip cluster
					if !diffRes {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to apply/delete/read resources
func (c *Context) run(ctx context.Context, g *errgroup.Group, op RsyncOperation, e RsyncEvent) error {

	if op == OpApply {
		// Setup dependency before starting app and cluster threads
		// Only for Apply for now
		for _, a := range c.ca.AppOrder {
			app := a
			if len(c.ca.Apps[app].Dependency) > 0 {
				c.dm.AddDependency(app, c.ca.Apps[app].Dependency)
			}
		}
	}
	// Iterate over all the subapps and start go Routines per app
	for _, a := range c.ca.AppOrder {
		app := a
		// If marked to skip then no processing needed
		if c.ca.Apps[app].Skip {
			log.Info("Update Skipping App::", log.Fields{"App": app})
			// Reset bit and skip app
			t := c.ca.Apps[app]
			t.Skip = false
			continue
		}
		g.Go(func() error {
			err := c.runApp(ctx, g, op, app, e)
			if op == OpApply {
				// Notify dependency that app got deployed
				c.dm.NotifyAppliedStatus(app)
			}
			return err
		})
	}
	return nil
}

func (c *Context) runApp(ctx context.Context, g *errgroup.Group, op RsyncOperation, app string, e RsyncEvent) error {

	if op == OpApply {
		// Check if any dependency and wait for dependencies to be met
		if err := c.dm.WaitForDependency(ctx, app); err != nil {
			return err
		}
	}
	// Iterate over all clusters
	for _, cluster := range c.ca.Apps[app].Clusters {
		// If marked to skip then no processing needed
		if cluster.Skip {
			log.Info("Update Skipping Cluster::", log.Fields{"App": app, "cluster": cluster})
			// Reset bit and skip cluster
			cluster.Skip = false
			continue
		}
		cluster := cluster.Name
		g.Go(func() error {
			return c.runCluster(ctx, op, e, app, cluster)
		})
	}
	return nil
}

func (c *Context) runCluster(ctx context.Context, op RsyncOperation, e RsyncEvent, app, cluster string) error {
	log.Info(" runCluster::", log.Fields{"app": app, "cluster": cluster})
	namespace, level := c.acRef.GetNamespace()
	cl, err := c.con.GetClientProviders(app, cluster, level, namespace)
	if err != nil {
		log.Error("Error in creating client", log.Fields{"error": err, "cluster": cluster, "app": app})
		return err
	}
	defer cl.CleanClientProvider()
	// Start cluster watcher if there are resources to be watched
	// case like admin cloud has no resources
	if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
		err = cl.StartClusterWatcher()
		if err != nil {
			log.Error("Error starting Cluster Watcher", log.Fields{
				"error":   err,
				"cluster": cluster,
			})
			return err
		}
	}
	r := resProvd{app: app, cluster: cluster, cl: cl, context: *c}
	// Timer key
	key := app + depend.SEPARATOR + cluster
	switch e {
	case InstantiateEvent:
		// Apply config for the cluster if there are any resources to be applied
		if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
			err = cl.ApplyConfig(ctx, nil)
			if err != nil {
				return err
			}
		}
		// Check if delete of status tracker is scheduled, if so stop and delete the timer
		c.StopDeleteStatusCRTimer(key)
		// Based on the discussions in Helm handling of CRD's
		// https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"]) > 0 {
			// Create CRD Resources if they don't exist
			log.Info("Creating CRD Resources if they don't exist", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"]})
			_, err := r.handleResources(ctx, OpCreate, c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"])
			if err != nil {
				return err
			}
		}
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]) > 0 {
			log.Info("Installing preinstall hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]})
			// Add Status tracking
			if err := r.addStatusTracker(status.PreInstallHookLabel); err != nil {
				return err
			}
			// Install Preinstall hooks with wait
			_, err := r.handleResourcesWithWait(ctx, op, c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"])
			if err != nil {
				r.deleteStatusTracker(status.PreInstallHookLabel)
				return err
			}
			// Delete Status tracking, will be added after the main resources are added
			r.deleteStatusTracker(status.PreInstallHookLabel)
			log.Info("Done Installing preinstall hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]})
		}
		// Install main resources without wait
		log.Info("Installing main resources", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
		i, err := r.handleResources(ctx, op, c.ca.Apps[app].Clusters[cluster].ResOrder)
		// handle status tracking before exiting if at least one resource got handled
		if i > 0 {
			// Add Status tracking
			r.addStatusTracker("")
		}
		if err != nil {
			log.Error("Error installing resources for app", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
			return err
		}
		log.Info("Done Installing main resources", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]) > 0 {
			log.Info("Installing Post-install Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]})
			// Install Postinstall hooks with wait
			_, err = r.handleResourcesWithWait(ctx, op, c.ca.Apps[app].Clusters[cluster].Dependency["post-install"])
			if err != nil {
				return err
			}
			log.Info("Done Installing Post-install Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]})
		}
	case TerminateEvent:
		// Apply Predelete hooks with wait
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]) > 0 {
			log.Info("Deleting pre-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]})
			_, err = r.handleResourcesWithWait(ctx, OpApply, c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"])
			if err != nil {
				return err
			}
			log.Info("Done Deleting pre-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]})
		}
		// Delete main resources without wait
		_, err = r.handleResources(ctx, op, c.ca.Apps[app].Clusters[cluster].ResOrder)
		if err != nil {
			return err
		}
		// Apply Postdelete hooks with wait
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]) > 0 {
			log.Info("Deleting post-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]})
			_, err = r.handleResourcesWithWait(ctx, OpApply, c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"])
			if err != nil {
				return err
			}
			log.Info("Done Deleting post-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]})
		}
		var rl []string
		// Delete all hook resources also
		// Ignore errors - There can be errors if the hook resources are not applied
		// like rollback hooks and test hooks
		for _, d := range c.ca.Apps[app].Clusters[cluster].Dependency {
			rl = append(rl, d...)
		}
		// Ignore errors
		_, _ = r.handleResources(ctx, op, rl)

		// Delete config for the cluster if applied
		if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
			err = cl.DeleteConfig(ctx, nil)
			if err != nil {
				return err
			}
		}
		// Check if delete of status tracker is scheduled, if so stop and delete the timer
		// before scheduling a new one
		c.StopDeleteStatusCRTimer(key)
		timer := ScheduleDeleteStatusTracker(c.statusAcID, app, cluster, level, namespace, c.con)
		c.UpdateDeleteStatusCRTimer(key, timer)
	case UpdateEvent, UpdateDeleteEvent:
		// Update and Rollback hooks are not supported at this time
		var rl []string
		// Find resources to handle based on skip bit
		resOrder := c.ca.Apps[app].Clusters[cluster].ResOrder
		for _, res := range resOrder {
			// If marked to skip then no processing needed
			if !c.ca.Apps[app].Clusters[cluster].Resources[res].Skip {
				rl = append(rl, res)
			}
		}
		// Handle main resources without wait
		_, err = r.handleResources(ctx, op, rl)
		if err != nil {
			return err
		}
		// Add Status tracking if not already applied for the cluster
		if op == OpApply {
			r.addStatusTracker("")
		}
	}
	return nil
}

// Schedule delete status tracker to run after 2 mins
// This gives time for delete status to be recorded in the monitor CR
func ScheduleDeleteStatusTracker(acID, app, cluster, level, namespace string, con Connector) *time.Timer {

	DurationOfTime := time.Duration(120) * time.Second
	label := acID + "-" + app
	b, err := status.GetStatusCR(label, "")
	if err != nil {
		log.Error("Failed to get status CR for deleting", log.Fields{"error": err, "label": label})
		return &time.Timer{}
	}

	f := func() {
		cl, err := con.GetClientProviders(app, cluster, level, namespace)
		if err != nil {
			log.Error("Error in creating client", log.Fields{"error": err, "cluster": cluster, "app": app})
			return
		}
		defer cl.CleanClientProvider()
		if err = cl.DeleteStatusCR(label, b); err != nil {
			log.Info("Failed to delete res", log.Fields{"error": err, "app": app, "label": label})
			return
		}
	}
	// Schedule for running at a later time
	handle := time.AfterFunc(DurationOfTime, f)
	return handle
}
