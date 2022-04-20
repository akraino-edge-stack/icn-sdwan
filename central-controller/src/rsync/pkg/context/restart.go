// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package context

/*
restart.go contains all the functions required for resuming the RYSNC once it restarts
*/

import (
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/connector"
)

const prefix string = "/activecontext/"

// RecordActiveContext shall insert into contextDB a key and value like /activecontext/99999999888/->99999999888. 99999999888 is sample AppcontextID
// It shall take in activeContextID
func RecordActiveContext(acID string) (bool, error) {

	exists, _ := ifContextIDActive(acID)
	if exists {
		logutils.Info("ContextID already active", logutils.Fields{"acID": acID})
		return false, nil
	}

	k := prefix + acID + "/"
	err := contextdb.Db.Put(k, acID)
	if err != nil {
		logutils.Info("Error saving the active contextID", logutils.Fields{"err": err.Error(), "contextID": acID})
		return false, pkgerrors.Errorf("Error:: %s saving contextID:: %s", err.Error(), acID)
	}

	return true, nil
}

// GetAllActiveContext shall return all the active contextIDs
func GetAllActiveContext() ([]string, error) {
	aContexts, err := contextdb.Db.GetAllKeys(prefix)
	if err != nil {
		logutils.Info("No active context handles for restart", logutils.Fields{})
		return nil, nil
	}
	aCtxIDs := make([]string, len(aContexts))
	for i, s := range aContexts {
		str := fmt.Sprintf("%v", s)
		key := strings.Split(str, "/")
		if len(key) == 4 && key[1] == "activecontext" {
			aCtxIDs[i] = key[2]
		}
	}
	return aCtxIDs, nil
}

// ifContextIDActive takes in a contextID and checks if its active or not
func ifContextIDActive(acID string) (bool, error) {

	k := prefix + acID + "/"
	var value string
	err := contextdb.Db.Get(k, &value)
	if err != nil {
		return false, nil
	}

	return true, nil

}

// DeleteActiveContextRecord deletes an active contextID
func DeleteActiveContextRecord(acID string) (bool, error) {
	exists, _ := ifContextIDActive(acID)
	if !exists {
		logutils.Info("ContextID not active", logutils.Fields{"acID": acID})
		return false, pkgerrors.Errorf("Error deleting, contextID:: %s not active", acID)
	}
	k := prefix + acID + "/"
	err := contextdb.Db.Delete(k)
	if err != nil {
		logutils.Info("Error deleting the contextID", logutils.Fields{"acID": acID, "Error": err.Error()})
		return false, pkgerrors.Errorf("Error:: %s deleting contextID:: %s ", err.Error(), acID)
	}
	return true, nil
}

// RestoreActiveContext shall be called everytime the rsync restarts.
// It makes sure that the AppContexts which were in active state before rsync
// got cancelled, are restored and queued up again for processing.
func RestoreActiveContext() error {
	logutils.Info("Restoring active context....", logutils.Fields{})
	acIDs, err := GetAllActiveContext()
	if err != nil {
		logutils.Info("Error getting all active contextIDs while restoring", logutils.Fields{})
		return nil
	}

	logutils.Info("Total active contexts to be restored", logutils.Fields{"Total active contextIDs to be restored": len(acIDs)})

	for _, acID := range acIDs {
		con := connector.NewProvider(acID)
		err = RestartAppContext(acID, &con)
		if err != nil {
			logutils.Info("Error in restoring active contextID in HandleAppContext", logutils.Fields{"acID": acID, "Error": err})
			return nil
		}
	}
	return nil
}
