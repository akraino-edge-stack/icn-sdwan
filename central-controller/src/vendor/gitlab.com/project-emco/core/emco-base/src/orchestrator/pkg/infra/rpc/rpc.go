// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package rpc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/testdata"
)

type ContextUpdateRequest interface {
}

type ContextUpdateResponse interface {
}

type InstallAppRequest interface {
}

type InstallAppResponse interface {
}

type rpcInfo struct {
	conn *grpc.ClientConn
	host string
	port int
}

var mutex = &sync.Mutex{}
var rpcConnections = make(map[string]rpcInfo)

// https://github.com/grpc/grpc/blob/master/doc/connectivity-semantics-and-api.md
func isGoodState(state connectivity.State) bool {
	return state == connectivity.Ready || state == connectivity.Idle
}

func waitForReady(conn *grpc.ClientConn) (connectivity.State, bool) {
	state := conn.GetState()
	if isGoodState(state) {
		return state, true
	}

	waitCfg := time.Duration(config.GetConfiguration().GrpcConnReadyTime)
	waitTime := waitCfg * time.Millisecond
	log.Info("Grpc conn in bad state, will wait...",
		log.Fields{"state": state, "waitTime": waitTime})

	// The wait is done under mutex. TODO This may need a revisit.
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()
	conn.ResetConnectBackoff() // wake up subchannels in transient failure, if any
	changed := conn.WaitForStateChange(ctx, state)

	state = conn.GetState()
	if changed && isGoodState(state) {
		log.Info("Grpc conn moved to good state", log.Fields{"state": state})
		return state, true
	}
	return state, false
}

// GetRpcConn is used by RPC client code which needs the connection for a
// given controller for doing RPC calls with that controller.
func GetRpcConn(name string) *grpc.ClientConn {
	mutex.Lock()
	defer mutex.Unlock()
	val, ok := rpcConnections[name]
	if !ok {
		log.Error("GetRpcConn: no Grpc connection available.", log.Fields{"name": name})
		return nil
	}

	state := val.conn.GetState()
	log.Info("GetRpcConn: RPC connection info", log.Fields{"name": name, "conn": val.conn, "host": val.host, "port": val.port, "conn-state": state.String()})

	state, goodConn := waitForReady(val.conn)
	if goodConn {
		return val.conn
	}

	log.Error("GetRpcConn: Bad gRPC connection", log.Fields{"name": name, "conn": val.conn, "host": val.host, "port": val.port, "conn-state": state.String()})
	return nil
}

// UpdateRpcConn initializes, reuses or updates a grpc connection.
// It does not guarantee that the conn is in a good state.
func UpdateRpcConn(name, host string, port int) {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		if val.host == host && val.port == port { // reuse cached conn
			log.Info("UpdateRpcConn: found connection", log.Fields{
				"name":       name,
				"conn":       val.conn,
				"host":       val.host,
				"port":       val.port,
				"conn-state": val.conn.GetState().String()})

			// No need to waitForReady(): caller is not using the conn now.
			return
		}
		// mismatch with cached connect info: close conn
		log.Info("UpdateRpcConn: closing RPC connection due to mismatch", log.Fields{
			"Server":   name,
			"Old Host": val.host,
			"Old Port": val.port,
			"New Host": host,
			"New Port": port,
		})
		err := val.conn.Close()
		if err != nil {
			log.Warn("UpdateRpcConn: error closing RPC connection", log.Fields{
				"Server": name,
				"Host":   val.host,
				"Port":   val.port,
				"Error":  err,
			})
		}
		// fallthrough to conn creation
	}

	// Either no cached conn or it is stale: create conn, update rpcConnection
	conn, err := createClientConn(host, port)
	if err != nil {
		log.Error("UpdateRpcConn: failed to create grpc connection", log.Fields{
			"Error": err,
			"Host":  host,
			"Port":  port,
		})
		delete(rpcConnections, name)
		return
	}
	rpcConnections[name] = rpcInfo{conn: conn, host: host, port: port}
	log.Info("UpdateRpcConn: added RPC Client connection", log.Fields{
		"Controller":    name,
		"rpcConnection": fmt.Sprintf("%v", rpcConnections[name])})
}

// CloseAllRpcConn closes all connections
func CloseAllRpcConn() {
	mutex.Lock()
	defer mutex.Unlock()
	log.Info("CloseAllRpcConn..", nil)
	for k, v := range rpcConnections {
		err := v.conn.Close()
		if err != nil {
			log.Warn("Error closing RPC connection", log.Fields{
				"Server": k,
				"Host":   v.host,
				"Port":   v.port,
				"Error":  err,
			})
		}
	}
}

// RemoveRpcConn closes the connection and removes from map
func RemoveRpcConn(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		err := val.conn.Close()
		if err != nil {
			log.Warn("Error closing RPC connection", log.Fields{
				"Server": name,
				"Host":   val.host,
				"Port":   val.port,
				"Error":  err,
			})
		}
		delete(rpcConnections, name)
	}
}

// createConn creates the Rpc Client Connection
func createClientConn(Host string, Port int) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	serverAddr := Host + ":" + strconv.Itoa(Port)
	serverNameOverride := config.GetConfiguration().GrpcServerNameOverride

	tls := strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable")

	caFile := config.GetConfiguration().GrpcCAFile

	if tls {
		if caFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, serverNameOverride)
		if err != nil {
			log.Error("Failed to create TLS credentials", log.Fields{
				"Error": err,
				"Host":  Host,
				"Port":  Port,
			})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	dialOpts := getGrpcDialOpts()
	opts = append(opts, dialOpts...)
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		pkgerrors.Wrap(err, "Grpc Connection Initialization failed with error")
	}

	return conn, err
}

func getGrpcDialOpts() []grpc.DialOption {
	// ConnTimeout is used for both ping period and ping timeout because
	// both relate to ready -> not-ready transition.
	pingTime := time.Duration(config.GetConfiguration().GrpcConnTimeout)
	pingTimeout := time.Duration(config.GetConfiguration().GrpcConnTimeout)

	// Add keepalive pings with timeout
	keepaliveParams := keepalive.ClientParameters{
		Time:    pingTime * time.Millisecond,
		Timeout: pingTimeout * time.Millisecond,
		// no pings without in-flight calls
		PermitWithoutStream: false, // default=false but we make it explicit
	}

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepaliveParams),
	}

	return opts
}
