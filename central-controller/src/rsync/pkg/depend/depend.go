// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package depend

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

type DependManager struct {
	acID string
	// Per App Ready channels to notify
	readyCh map[string][]appData
	// Per App Deploy channels to notify
	deployedCh map[string][]appData
	// Per app channels to wait on
	appCh map[string][]chan struct{}
	// Single Resource succeed channels
	// Map of app+cluster inside map is of resource
	resCh map[string]map[string]resData
	sync.RWMutex
}

type resData struct {
	res string
	// Chan to report if resource is ready/success
	ch chan struct{}
}

type appData struct {
	app string
	crt types.Criteria
	// Chan to report if app meets Criteria
	ch chan struct{}
}
type DependManagerList struct {
	dm map[string]*DependManager
	sync.RWMutex
}

var dmList DependManagerList

func init() {
	dmList.dm = make(map[string]*DependManager)
}

const SEPARATOR = "+"

// New Manager for acID
func NewDependManager(acID string) *DependManager {
	d := DependManager{
		acID: acID,
	}
	d.deployedCh = make(map[string][]appData)
	d.readyCh = make(map[string][]appData)
	d.appCh = make(map[string][]chan struct{})
	d.resCh = make(map[string]map[string]resData)
	dmList.Lock()
	dmList.dm[acID] = &d
	dmList.Unlock()
	return &d
}

// Function registers an app for dependency
func (dm *DependManager) AddDependency(app string, dep map[string]*types.Criteria) error {

	dm.Lock()
	defer dm.Unlock()

	log.Info("AddDependency", log.Fields{"app": app, "dep": dep})
	for d, c := range dep {
		depLabel := d
		ch := make(chan struct{}, 1)
		data := appData{app: app, crt: *c, ch: ch}
		if c.OpStatus == types.OpStatusReady {
			dm.readyCh[depLabel] = append(dm.readyCh[depLabel], data)
		} else if c.OpStatus == types.OpStatusDeployed {
			dm.deployedCh[depLabel] = append(dm.deployedCh[depLabel], data)
		} else {
			// Ignore it
			continue
		}
		// Add all channels to per app list
		dm.appCh[app] = append(dm.appCh[app], ch)
	}
	return nil
}

// WaitResourceDependency waits for the resource to be Success
func (dm *DependManager) WaitResourceDependency(ctx context.Context, app, cluster, name string) error {

	key := app + SEPARATOR + cluster
	ch := make(chan struct{}, 1)
	dm.Lock()
	// Add channels to wait list
	if len(dm.resCh[key]) == 0 {
		dm.resCh[key] = make(map[string]resData)
	}
	dm.resCh[key][name] = resData{res: name, ch: ch}
	dm.Unlock()
	// Remove the channel from the list on function return
	defer func() {
		dm.Lock()
		delete(dm.resCh[key], name)
		dm.Unlock()
	}()
	// Start with wait time of 1 sec and then change it to
	// 30 secs to avoid missing first notification
	waitTime := 1
	for {
		select {
		// Wait for wait time before checking if resource is ready
		case <-time.After(time.Duration(waitTime) * time.Second):
			// Context is canceled
			if ctx.Err() != nil {
				return ctx.Err()
			}
			waitTime = 30
			b := dm.GetResourceReadyStatus(app, cluster, name)
			if b {
				return nil
			} else {
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			// Channel notified and resource is ready
			return nil
		}
	}
}
func (dm *DependManager) ClearChannels(app string) {
	dm.Lock()
	// Find all the entries for the app in deployedCh and readyCh
	for _, x := range []map[string][]appData{dm.deployedCh, dm.readyCh} {
		for _, d := range x {
			for i := len(d) - 1; i >= 0; i-- {
				if d[i].app == app {
					d = append(d[:i], d[i+1:]...)
				}
			}
		}
	}
	delete(dm.appCh, app)
	dm.Unlock()
}

//WaitForDependency waits for all dependecies to be met for an app
func (dm *DependManager) WaitForDependency(ctx context.Context, app string) error {

	var cases []reflect.SelectCase
	dm.RLock()
	if len(dm.appCh[app]) <= 0 {
		dm.RUnlock()
		return nil
	}
	log.Info("WaitForDependency", log.Fields{"app": app})
	// Add the case for ctx done
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	})
	for _, ch := range dm.appCh[app] {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}
	dm.RUnlock()
	// When wait is done remove all channels form the Dependency arrays
	defer dm.ClearChannels(app)
	num := len(cases) - 1
	// Wait for all channels to be available
	for i := 0; i < num; i++ {
		// Wait for all channels
		index, _, _ := reflect.Select(cases)
		switch index {
		case 0:
			// case <- ctx.Done()
			return ctx.Err()
		default:
			// Delete the channel from list
			log.Info("WaitForDependency: Coming out of wait", log.Fields{"app": app, "index": index})
			// Some channel is done, remove it from the list
			cases = append(cases[:index], cases[index+1:]...)
			continue
		}
	}
	return nil
}

func (dm *DependManager) NotifyAppliedStatus(app string) {
	// Read deployed channel
	dm.RLock()
	dl := dm.deployedCh[app]
	dm.RUnlock()
	dm.NotifyStatus(dl)
}

func (dm *DependManager) NotifyReadyStatus(app string) {
	// Read ready channel
	dm.RLock()
	dl := dm.readyCh[app]
	dm.RUnlock()
	dm.NotifyStatus(dl)
}

func (dm *DependManager) NotifyStatus(dl []appData) {
	for _, d := range dl {
		// Wait and notify channels
		go func(d appData) {
			if d.crt.Wait != 0 {
				time.Sleep(time.Duration(d.crt.Wait) * time.Second)
			}
			d.ch <- struct{}{}
		}(d)
	}
}

// Inform waiting threads of App Ready
func ResourcesReady(acID, app, cluster string) {

	dmList.RLock()
	// Check if AppContext has dependency
	dm, ok := dmList.dm[acID]
	if !ok {
		dmList.RUnlock()
		return
	}
	dmList.RUnlock()
	key := app + SEPARATOR + cluster
	// If no app is waiting for ready status of the app
	// Not further processing needed
	dm.RLock()
	if len(dm.readyCh[app]) == 0 && len(dm.resCh[key]) == 0 {
		dm.RUnlock()
		return
	}
	length := len(dm.readyCh[app])
	dm.RUnlock()
	if length > 0 {
		// Inform waiting apps
		go func() {
			// Check in appContext if app is ready on all clusters inform ready status
			acUtils, err := utils.NewAppContextReference(acID)
			if err != nil {
				return
			}
			if acUtils.CheckAppReadyOnAllClusters(app) {
				// Notify the apps waiting for the app to be ready
				dm.NotifyReadyStatus(app)
			}
		}()
	}
	dm.RLock()
	res := dm.resCh[app]
	dm.RUnlock()
	if len(res) > 0 {
		go func() {
			for _, r := range res {
				b := dm.GetResourceReadyStatus(app, cluster, r.res)
				if b {
					// If succeded inform the waiting channel
					r.ch <- struct{}{}
				}
			}
		}()
	}
}

func (dm *DependManager) GetResourceReadyStatus(app, cluster, res string) bool {
	var resStatus bool
	acUtils, err := utils.NewAppContextReference(dm.acID)
	if err != nil {
		return false
	}
	result := strings.SplitN(res, "+", 2)
	if len(result) != 2 {
		log.Error("Invalid resource name format::", log.Fields{"res": res})
		return false
	}
	// In case of Pod and Job Success state is used for Hook readiness
	if result[1] == "Pod" || result[1] == "Job" {
		resStatus = acUtils.GetResourceReadyStatus(app, cluster, res, string(types.SuccessStatus))
	} else {
		resStatus = acUtils.GetResourceReadyStatus(app, cluster, res, string(types.ReadyStatus))
	}
	return resStatus
}
