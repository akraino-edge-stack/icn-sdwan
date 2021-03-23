// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package readynotifyserver

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
	pkgerrors "github.com/pkg/errors"
)

type readyNotifyServer struct {
	name         string
	alertStrm    map[string]pb.ReadyNotify_AlertServer
	appContextId map[string]string
}

func (s *readyNotifyServer) Alert(topic *pb.Topic, stream pb.ReadyNotify_AlertServer) error {
	client := topic.GetClientName()
	appContextId := topic.GetAppContext()
	log.Printf("[ReadyNotify gRPC] Received an Alert subscription request (%s, %s)", client, appContextId)
	s.alertStrm[client] = stream
	s.appContextId[client] = appContextId
	for {
		time.Sleep(3 * time.Second)
		for k, v := range s.alertStrm {
			appContextId = s.appContextId[k]
			logutils.Info("[ReadyNotify gRPC] Checking cluster statuses", logutils.Fields{"appContextId": appContextId})
			err := checkAppContext(appContextId)
			if err == nil {
				logutils.Info("[ReadyNotify gRPC] All clusters' certificates have been issued", logutils.Fields{"appContextId": appContextId})
				err = v.Send(&pb.Notification{AppContext: appContextId})
				if err != nil {
					logutils.Error("[ReadyNotify gRPC] Alert notification back to client failed to be sent", logutils.Fields{"err": err})
				} else {
					// we're done with this appcontext ID, remove it from map (safe in Go within for loop)
					delete(s.alertStrm, k)
				}
			} else {
				logutils.Info("[ReadyNotify gRPC] Not all clusters' certificates have been issued, yet", logutils.Fields{"appContextId": appContextId})
			}
		}
	}
}

// checkAppContext checks whether the LC from the provided appcontext has had certificates issued by every cluster
func checkAppContext(appContextId string) error {
	// Get the contextId from the label (id)
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		logutils.Error("AppContext not found", logutils.Fields{"appContextId": appContextId})
		return err
	}

	appsOrder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		clusterNames, err := ac.GetClusterNames(app)
		if err != nil {
			return err
		}
		// iterate over all clusters of appcontext
		for k := 0; k < len(clusterNames); k++ {
			chandle, err := ac.GetClusterHandle(app, clusterNames[k])
			if err != nil {
				logutils.Info("Error getting cluster handle", logutils.Fields{"clusterNames[k]": clusterNames[k]})
				return err
			}
			// Get the handle for the cluster status object
			handle, _ := ac.GetLevelHandle(chandle, "status")

			clusterStatus, err := ac.GetValue(handle)
			if err != nil {
				logutils.Error("Couldn't fetch cluster status from its handle", logutils.Fields{"handle": handle})
				return err
			}
			// detect if certificate has been issued - assumes K8s base64-encoded PEM certificate
			if strings.Contains(clusterStatus.(string), "certificate\":\"LS0t") {
				logutils.Info("Cluster status contains the certificate", logutils.Fields{})
				return nil
			} else {
				logutils.Info("Cluster status doesn't contain the certificate yet", logutils.Fields{})
			}
		}
	}
	return pkgerrors.New("Certificates not issued yet")
}

// NewReadyNotifyServer exported
func NewReadyNotifyServer() *readyNotifyServer {
	s := &readyNotifyServer{
		name:         "readyNotifyServer",
		alertStrm:    make(map[string]pb.ReadyNotify_AlertServer),
		appContextId: make(map[string]string),
	}
	return s
}
