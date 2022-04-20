// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"encoding/json"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	types "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

type AppContextQueueUtils struct {
	ac appcontext.AppContext
}

// GetAppContextQueue shall return the AppContextQueue
func (aq *AppContextQueueUtils) GetAppContextQueue() (types.AppContextQueue, error) {

	h, err := aq.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error getting CompAppHandle", log.Fields{"err": err})
		return types.AppContextQueue{}, err
	}

	qKey := types.AppContextEventQueueKey

	qh, err := aq.ac.GetLevelHandle(h, qKey)
	if err != nil {
		log.Info("Error in getting the AppContextQueue Level handle", log.Fields{"err": err})
		return types.AppContextQueue{}, err
	}
	if qh != nil {
		v, err := aq.ac.GetValue(qh)
		if err != nil {
			log.Error("Error getting vale for the AppQ handle", log.Fields{"err": err})
		}

		acQ := types.AppContextQueue{}
		js, err := json.Marshal(v)
		if err != nil {
			log.Error("Marshal Error in GetAppContextQueue", log.Fields{"err": err})
			return types.AppContextQueue{}, err
		}
		err = json.Unmarshal(js, &acQ)
		if err != nil {
			log.Error("UnMarshal Error in GetAppContextQueue", log.Fields{"err": err})
			return types.AppContextQueue{}, err
		}
		return acQ, nil

	}
	log.Info("AppContextQueue Level handle is NULL", log.Fields{})
	return types.AppContextQueue{}, err
}

// GetAppContextQueuePeek shall return the String value at the peak of the AppContextQueue
func (aq *AppContextQueueUtils) GetAppContextQueuePeek() (types.AppContextQueueElement, error) {
	acQ, err := aq.GetAppContextQueue()
	if err != nil {
		log.Error("Error in getting AppContextQueue", log.Fields{"err": err})
		return types.AppContextQueueElement{}, err
	}
	q := acQ.AcQueue
	if len(q) == 0 {
		return types.AppContextQueueElement{}, pkgerrors.Errorf("No element in AppContextQueue")
	}

	return q[0], nil
}

// GetAppContextQueueLength shall return the length of the AppContextQueue
func (aq *AppContextQueueUtils) GetAppContextQueueLength() (int, error) {
	acQ, err := aq.GetAppContextQueue()
	if err != nil {
		log.Error("Error in getting AppContextQueue", log.Fields{"err": err})
		return 0, err
	}
	q := acQ.AcQueue
	if len(q) == 0 {
		log.Info("The length of AppContextQueue is 0", log.Fields{"err": err})
		return 0, nil
	}
	return len(q), nil
}

// Enqueue shall append new Q-Elemenet in the string format
func (aq *AppContextQueueUtils) Enqueue(qElement types.AppContextQueueElement) (bool, error) {

	acQ, err := aq.GetAppContextQueue()
	if err != nil {
		log.Info("No existing AppContextQueue, Creating AppContextQueue", log.Fields{"err": err})
		return aq.CreateQueue(qElement)
	}
	if len(acQ.AcQueue) >= 1 {
		q := acQ.AcQueue
		q = append(q, qElement)
		return aq.UpdateQueue(q)
	}
	return false, nil
}

func (aq *AppContextQueueUtils) UpdateQueue(q []types.AppContextQueueElement) (bool, error) {
	h, err := aq.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error in getting CompApp handle in UpdateQueue", log.Fields{"err": err})
		return false, err
	}

	qKey := types.AppContextEventQueueKey

	qHandle, err := aq.ac.GetLevelHandle(h, qKey)
	if err != nil {
		log.Error("Error in getting Qhandle", log.Fields{"err": err})
		return false, err
	}
	acQ := types.AppContextQueue{AcQueue: q}
	err = aq.ac.UpdateValue(qHandle, acQ)
	if err != nil {
		log.Error("Error in updating Qhandle", log.Fields{"err": err})
		return false, err
	}

	return true, nil
}

func (aq *AppContextQueueUtils) CreateQueue(qElement types.AppContextQueueElement) (bool, error) {
	h, err := aq.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error in getting CompApp handle in CreateQueue", log.Fields{"err": err})
		return false, err
	}
	var q []types.AppContextQueueElement
	q = append(q, qElement)
	acQ := types.AppContextQueue{AcQueue: q}

	qKey := types.AppContextEventQueueKey
	qHandle, err := aq.ac.AddLevelValue(h, qKey, acQ)
	if err != nil {
		log.Error("Error in Adding AppContextQueue Level", log.Fields{"err": err})
		return false, err
	}
	qhandle := fmt.Sprintf("%v", qHandle)
	log.Info("AppContextQueue created :: Qhandle :: ", log.Fields{"qhandle": qhandle})
	return true, nil
}
func (aq *AppContextQueueUtils) FindFirstPending() (int, types.AppContextQueueElement) {

	q, err := aq.GetAppContextQueue()
	if err != nil {
		return -1, types.AppContextQueueElement{}
	}
	for i, v := range q.AcQueue {
		if v.Status == "Pending" {
			return i, v
		}
	}
	return -1, types.AppContextQueueElement{}
}
func (aq *AppContextQueueUtils) UpdateStatus(index int, status string) error {
	acQ, err := aq.GetAppContextQueue()
	if err != nil {
		log.Error("Error in getting AppContextQueue", log.Fields{"err": err})
		return err
	}
	q := acQ.AcQueue
	if index >= len(q) || index < 0 {
		return pkgerrors.Errorf("Invalid index AppContextQueue")
	}
	q[index].Status = status
	_, err = aq.UpdateQueue(q)
	return err
}
