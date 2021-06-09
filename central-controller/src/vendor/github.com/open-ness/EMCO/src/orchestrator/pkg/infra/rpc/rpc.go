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

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
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

// GetRpcConn is used by RPC client code which needs the connection for a
// given controller for doing RPC calls with that controller.
func GetRpcConn(name string) *grpc.ClientConn {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		log.Info("GetRpcConn .. RPC intermediate-connection info", log.Fields{"name": name, "conn": val.conn, "host": val.host, "port": val.port, "conn-state": val.conn.GetState().String()})

		if val.conn.GetState() == connectivity.TransientFailure {
			val.conn.ResetConnectBackoff()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if !val.conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
				log.Warn("Error re-establishing RPC connection", log.Fields{
					"Server":    name,
					"Host":      val.host,
					"Port":      val.port,
					"ConnState": val.conn.GetState().String(),
				})
				return nil
			}
		}

		if val.conn.GetState() == connectivity.Connecting {
			ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
			defer cancel()
			if !val.conn.WaitForStateChange(ctx, connectivity.Connecting) {
				log.Warn("GetRpcConn .. Error establishing RPC connection .. ClientConn is still in CONNECTING", log.Fields{
					"Server":    name,
					"Host":      val.host,
					"Port":      val.port,
					"ConnState": val.conn.GetState().String(),
				})
				return nil
			}
		}

		if val.conn.GetState() == connectivity.TransientFailure {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if !val.conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
				log.Warn("Error establishing RPC connection", log.Fields{
					"Server":    name,
					"Host":      val.host,
					"Port":      val.port,
					"ConnState": val.conn.GetState().String(),
				})
				return nil
			}
		}

		log.Warn("GetRpcConn .. RPC final-connection info", log.Fields{"name": name, "conn": val.conn, "host": val.host, "port": val.port, "conn-state": val.conn.GetState().String()})
		return val.conn
	}
	return nil
}

// UpdateRpcConn ... Update RPC connections
func UpdateRpcConn(name, host string, port int) {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		// close connection if mismatch in update vs cached connect info
		if val.host != host || val.port != port {
			log.Info("Closing RPC connection due to mismatch", log.Fields{
				"Server":   name,
				"Old Host": val.host,
				"Old Port": val.port,
				"New Host": host,
				"New Port": port,
			})
			err := val.conn.Close()
			if err != nil {
				log.Error("Error closing RPC connection", log.Fields{
					"Server": name,
					"Host":   val.host,
					"Port":   val.port,
					"Error":  err,
				})
			}
		} else {
			if val.conn.GetState() == connectivity.TransientFailure {
				val.conn.ResetConnectBackoff()
			}
			return
		}
	}
	// connect and update rpcConnection list - for new or modified connection
	conn, err := createClientConn(host, port)
	if err != nil {
		log.Error("Failed to create RPC Client connection", log.Fields{
			"Error": err,
		})
		delete(rpcConnections, name)
	} else {
		rpcConnections[name] = rpcInfo{
			conn: conn,
			host: host,
			port: port,
		}
		log.Info("Added RPC Client connection", log.Fields{"Controller": name, "conn": conn, "rpcConnection": fmt.Sprintf("%v", rpcConnections[name])})
	}
	log.Info("UpdateRpcConn .. end", log.Fields{"conn": conn, "name": name, "host": host, "port": port})
}

// CloseAllRpcConn closes all connections
func CloseAllRpcConn() {
	mutex.Lock()
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
	mutex.Unlock()
}

// RemoveRpcConn closes the connection and removes from map
func RemoveRpcConn(name string) {
	mutex.Lock()
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
	mutex.Unlock()
}

// createConn creates the Rpc Client Connection
func createClientConn(Host string, Port int) (*grpc.ClientConn, error) {
	var err error
	var tls bool
	var opts []grpc.DialOption

	serverAddr := Host + ":" + strconv.Itoa(Port)
	serverNameOverride := config.GetConfiguration().GrpcServerNameOverride

	if strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable") {
		tls = true
	} else {
		tls = false
	}

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

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		pkgerrors.Wrap(err, "Grpc Client Initialization failed with error")
	}

	return conn, err
}
