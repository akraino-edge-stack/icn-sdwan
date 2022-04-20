// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateclient

import (
	"context"
	"time"

	contextpb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	pkgerrors "github.com/pkg/errors"
)

// InvokeContextUpdate will make the grpc call to the specified controller
// The controller will take the specified intentName and update the AppContext
// appropriatly based on its operation as a placement or action controller.
func InvokeContextUpdate(controllerName, intentName, appContextId string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var updateRes *contextpb.ContextUpdateResponse

	timeout := time.Duration(config.GetConfiguration().GrpcCallTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Millisecond)
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
