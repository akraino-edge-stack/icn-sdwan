// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package client

import (
	"context"
	"sync"
	"time"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	installpb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/installapp"
	pkgerrors "github.com/pkg/errors"
)

const rsyncName = "rsync"

/*
RsyncInfo consists of rsyncName, hostName and portNumber.
*/
type RsyncInfo struct {
	RsyncName  string
	hostName   string
	portNumber int
}

var rsyncInfo RsyncInfo
var mutex = &sync.Mutex{}

type _testvars struct {
	UseGrpcMock   bool
	InstallClient installpb.InstallappClient
}

var Testvars _testvars

// InitRsyncClient initializes connctions to the Resource Synchronizer service
func initRsyncClient() bool {
	if (RsyncInfo{}) == rsyncInfo {
		mutex.Lock()
		defer mutex.Unlock()
		log.Error("RsyncInfo not set. InitRsyncClient failed", log.Fields{
			"Rsyncname":  rsyncInfo.RsyncName,
			"Hostname":   rsyncInfo.hostName,
			"PortNumber": rsyncInfo.portNumber,
		})
		return false
	}
	rpc.UpdateRpcConn(rsyncInfo.RsyncName, rsyncInfo.hostName, rsyncInfo.portNumber)
	return true
}

// NewRsyncInfo shall return a newly created RsyncInfo object
func NewRsyncInfo(rName, h string, pN int) RsyncInfo {
	mutex.Lock()
	defer mutex.Unlock()
	rsyncInfo = RsyncInfo{RsyncName: rName, hostName: h, portNumber: pN}
	return rsyncInfo

}

// InvokeInstallApp will make the grpc call to the resource synchronizer
// or rsync controller.
// rsync will deploy the resources in the app context to the clusters as
// prepared in the app context.
func InvokeInstallApp(appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var installRes *installpb.InstallAppResponse
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Unit test helper code
	if Testvars.UseGrpcMock {
		rpcClient = Testvars.InstallClient
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = appContextId
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
		return nil
	}

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		initRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
	}

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = appContextId
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("InstallApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if installRes.AppContextInstalled {
			log.Info("InstallApp Success", log.Fields{
				"AppContext": appContextId,
				"Message":    installRes.AppContextInstallMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("InstallApp Failed: %v", installRes.AppContextInstallMessage)
		}
	}
	return err
}

func InvokeUninstallApp(appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var uninstallRes *installpb.UninstallAppResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		initRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
	}

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		uninstallReq := new(installpb.UninstallAppRequest)
		uninstallReq.AppContext = appContextId
		uninstallRes, err = rpcClient.UninstallApp(ctx, uninstallReq)
		if err == nil {
			log.Info("Response from UninstappApp GRPC call", log.Fields{
				"Succeeded": uninstallRes.AppContextUninstalled,
				"Message":   uninstallRes.AppContextUninstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("UninstallApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if uninstallRes.AppContextUninstalled {
			log.Info("UninstallApp Success", log.Fields{
				"AppContext": appContextId,
				"Message":    uninstallRes.AppContextUninstallMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("UninstallApp Failed: %v", uninstallRes.AppContextUninstallMessage)
		}
	}
	return err
}

func InvokeGetResource(appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var readAppContextRes *installpb.ReadAppContextResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		initRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
		if conn == nil {
			log.Error("[InvokeReadRq gRPC] connection error", log.Fields{"grpc-server": rsyncName})
			return pkgerrors.Errorf("[InvokeReadRq gRPC] connection error. grpc-server[%v]", rsyncName)
		}
	}

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		readReq := new(installpb.ReadAppContextRequest)
		readReq.AppContext = appContextId
		readAppContextRes, err = rpcClient.ReadAppContext(ctx, readReq)
		if err == nil {
			log.Info("Response from ReadAppContext GRPC call", log.Fields{
				"Succeeded": readAppContextRes.AppContextReadSuccessful,
				"Message":   readAppContextRes.AppContextReadMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("ReadAppContext Failed - Could not get ReadAppContext: %v", "rsync")
	}

	return nil
}
