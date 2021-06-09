// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	register "github.com/open-ness/EMCO/src/rsync/pkg/grpc"
	installpb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/installapp"
	"github.com/open-ness/EMCO/src/rsync/pkg/grpc/installappserver"
	readynotifypb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
	"github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotifyserver"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
	contextDb "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

func startGrpcServer() error {
	var tls bool

	if strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable") {
		tls = true
	} else {
		tls = false
	}
	certFile := config.GetConfiguration().GrpcServerCert
	keyFile := config.GetConfiguration().GrpcServerKey

	_, port := register.GetServerHostPort()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not listen to port: %v", err)
	}
	var opts []grpc.ServerOption
	if tls {
		if certFile == "" {
			certFile = testdata.Path("server.pem")
		}
		if keyFile == "" {
			keyFile = testdata.Path("server.key")
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Could not generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	
	grpcServer := grpc.NewServer(opts...)
	installpb.RegisterInstallappServer(grpcServer, installappserver.NewInstallAppServer())
	readynotifypb.RegisterReadyNotifyServer(grpcServer, readynotifyserver.NewReadyNotifyServer())

	log.Println("Starting rsync gRPC Server")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("rsync grpc server is not serving %v", err)
	}
	return err
}

func main() {

	rand.Seed(time.Now().UnixNano())

	// Initialize the mongodb
	err := db.InitializeDatabaseConnection("scc")
	if err != nil {
		log.Println("Unable to initialize mongo database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Initialize contextdb
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Println("Unable to initialize etcd database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	err = startGrpcServer()
	if err != nil {
		log.Fatalf("GRPC server failed to start")
	}
}
