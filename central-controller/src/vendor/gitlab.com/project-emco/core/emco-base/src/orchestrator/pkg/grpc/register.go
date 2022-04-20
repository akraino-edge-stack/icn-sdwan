// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package grpc

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	updatepb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/contextupdate"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	statusnotifypb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"google.golang.org/grpc"
)

func RegisterStatusNotifyService(grpcServer *grpc.Server, srv interface{}) {
	statusnotifypb.RegisterStatusNotifyServer(grpcServer, srv.(statusnotify.StatusNotifyServer))
}

func RegisterContextUpdateService(grpcServer *grpc.Server, srv interface{}) {
	updatepb.RegisterContextupdateServer(grpcServer, srv.(updatepb.ContextupdateServer))
}

func StartGrpcServer(defaultName, envName string, defaultPort int, registerFn func(*grpc.Server, interface{}), srv interface{}) error {
	port := getGrpcServerPort(defaultName, envName, defaultPort)

	log.Info("Starting gRPC on port", log.Fields{"Port": port})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error("Could not listen to gRPC port", log.Fields{"Error": err})
		return err
	}

	grpcServer := grpc.NewServer()
	registerFn(grpcServer, srv)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Error("gRPC server is not serving", log.Fields{"Error": err})
		return err
	}
	return err
}

func getGrpcServerPort(defaultName, envName string, defaultPort int) int {

	// expect name of this program to be in env the variable "{envName}_NAME" - e.g. ORCHESTRATOR_NAME="orchestrator"
	serviceName := os.Getenv(envName)
	if serviceName == "" {
		serviceName = defaultName
		log.Info("Using default name for service", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service port to be in env variable - e.g. ORCHESTRATOR_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = defaultPort
		log.Info("Using default port for gRPC controller", log.Fields{
			"Name": serviceName,
			"Port": port,
		})
	}
	return port
}
