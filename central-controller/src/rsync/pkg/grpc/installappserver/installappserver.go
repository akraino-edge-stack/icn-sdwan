// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package installappserver

import (
	"context"
	"encoding/json"
	con "github.com/open-ness/EMCO/src/rsync/pkg/context"
	"github.com/open-ness/EMCO/src/rsync/pkg/grpc/installapp"
	"log"
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

// NewInstallAppServer exported
func NewInstallAppServer() *installappServer {
	s := &installappServer{}
	return s
}
