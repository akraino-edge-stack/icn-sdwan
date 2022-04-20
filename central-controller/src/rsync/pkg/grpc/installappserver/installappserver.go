// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package installappserver

import (
	"context"
	"encoding/json"
	"log"

	con "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/installapp"
)

type installappServer struct {
	installapp.UnimplementedInstallappServer
}

func (cs *installappServer) InstallApp(ctx context.Context, req *installapp.InstallAppRequest) (*installapp.InstallAppResponse, error) {
	installAppReq, _ := json.Marshal(req)
	log.Println("GRPC Server received installAppRequest: ", string(installAppReq))

	// Try instantiate the comp app
	instca := con.CompositeAppContext{}
	err := instca.InstantiateComApp(req.GetAppContext())
	if err != nil {
		log.Println("Instantiation failed: " + err.Error())
		err := instca.TerminateComApp(req.GetAppContext())
		if err != nil {
			log.Println("Termination failed: " + err.Error())
		}
		return &installapp.InstallAppResponse{AppContextInstalled: false}, err
	}
	return &installapp.InstallAppResponse{AppContextInstalled: true}, nil
}

func (cs *installappServer) UninstallApp(ctx context.Context, req *installapp.UninstallAppRequest) (*installapp.UninstallAppResponse, error) {
	uninstallAppReq, _ := json.Marshal(req)
	log.Println("GRPC Server received uninstallAppRequest: ", string(uninstallAppReq))

	// Try terminating the comp app here
	instca := con.CompositeAppContext{}
	err := instca.TerminateComApp(req.GetAppContext())
	if err != nil {
		log.Println("Termination failed: " + err.Error())
		return &installapp.UninstallAppResponse{AppContextUninstalled: false}, err
	}

	return &installapp.UninstallAppResponse{AppContextUninstalled: true}, nil
}


func (cs *installappServer) ReadAppContext(ctx context.Context, req *installapp.ReadAppContextRequest) (*installapp.ReadAppContextResponse, error){
	readAppContext, _ := json.Marshal(req)
	log.Println("GRPC Server received ReadAppContext: ", string(readAppContext))

	// Try instantiate the comp app
	instca := con.CompositeAppContext{}
	err := instca.ReadComApp(req.GetAppContext())
	if err != nil {
		log.Println("Termination failed: " + err.Error())
		return &installapp.ReadAppContextResponse{AppContextReadSuccessful: false, AppContextReadMessage: "AppContext read failed"}, err
	}

	return &installapp.ReadAppContextResponse{AppContextReadSuccessful: true, AppContextReadMessage: "AppContext read successfully"}, nil
}

// NewInstallAppServer exported
func NewInstallAppServer() *installappServer {
	s := &installappServer{}
	return s
}
