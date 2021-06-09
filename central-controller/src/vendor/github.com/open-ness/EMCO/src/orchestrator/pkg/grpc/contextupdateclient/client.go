// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateclient

import (
	"context"
	"time"

	contextpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/contextupdate"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"
)

// InvokeContextUpdate will make the grpc call to the specified controller
// The controller will take the specified intentName and update the AppContext
// appropriatly based on its operation as a placement or action controller.
func InvokeContextUpdate(controllerName, intentName, appContextId string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var updateRes *contextpb.ContextUpdateResponse
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(controllerName)
	if conn != nil {
		rpcClient = contextpb.NewContextupdateClient(conn)
		updateReq := new(contextpb.ContextUpdateRequest)
		updateReq.AppContext = appContextId
		updateReq.IntentName = intentName
		updateRes, err = rpcClient.UpdateAppContext(ctx, updateReq)
	} else {
		return pkgerrors.Errorf("ContextUpdate Failed - Could not get ContextupdateClient: %v", controllerName)
	}

	if err == nil {
		if updateRes.AppContextUpdated {
			log.Info("ContextUpdate Passed", log.Fields{
				"Controller": controllerName,
				"Intent":     intentName,
				"AppContext": appContextId,
				"Message":    updateRes.AppContextUpdateMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("ContextUpdate Failed: %v", updateRes.AppContextUpdateMessage)
		}
	}
	return err
}
