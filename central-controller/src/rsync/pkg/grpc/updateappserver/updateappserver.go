// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package updateappserver

import (
	"context"
	"encoding/json"
	"log"

	con "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/updateapp"
)

type updateappServer struct {
	updateapp.UnimplementedUpdateappServer
}

func (cs *updateappServer) UpdateApp(ctx context.Context, req *updateapp.UpdateAppRequest) (*updateapp.UpdateAppResponse, error) {
	updateAppReq, _ := json.Marshal(req)
	log.Println("GRPC Server received UpdateAppRequest: ", string(updateAppReq))

	// Try updating the comp app
	instca := con.CompositeAppContext{}
	err := instca.UpdateComApp(req.GetUpdateFromAppContext(), req.GetUpdateToAppContext())
	if err != nil {
		log.Println("Updating the compApp failed: " + err.Error())
		return &updateapp.UpdateAppResponse{AppContextUpdated: false}, err
	}
	return &updateapp.UpdateAppResponse{AppContextUpdated: true}, nil
}

func (cs *updateappServer) RollbackApp(ctx context.Context, req *updateapp.RollbackAppRequest) (*updateapp.RollbackAppResponse, error) {
	updateAppReq, _ := json.Marshal(req)
	log.Println("GRPC Server received UpdateAppRequest: ", string(updateAppReq))

	// Try rollback for the comp app
	instca := con.CompositeAppContext{}
	err := instca.UpdateComApp(req.GetRollbackFromAppContext(), req.GetRollbackToAppContext())
	if err != nil {
		log.Println("Rollback for compApp failed: " + err.Error())
		return &updateapp.RollbackAppResponse{AppContextRolledback: false}, err
	}
	return &updateapp.RollbackAppResponse{AppContextRolledback: true}, nil
}

// NewInstallAppServer exported
func NewUpdateAppServer() *updateappServer {
	s := &updateappServer{}
	return s
}
