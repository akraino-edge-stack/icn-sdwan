// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2021 Intel Corporation

package readynotifyserver

import (
	"context"
	"log"
	"sync"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
)

// readyNotifyServer will be initialized by NewReadyNotifyServer() and
// its lifecycle is valid until all the clients unsubscribed the stream notification channel
type readyNotifyServer struct {
	name string
	// alertNotify contains the map of "appContextID" and "map of client name and stream server"
	// For Ex: map[12345687:map[[dtc:st1] [dcm:st2]] 456785369:map[[ncm:st3] [dtc:st4]]]
	alertNotify   map[string]map[string]pb.ReadyNotify_AlertServer
	streamChannel map[pb.ReadyNotify_AlertServer]chan int
	mutex         sync.Mutex
}

var notifServer *readyNotifyServer

// Alert gets notified when the subscriber subscribes for an appcontext event notification
func (s *readyNotifyServer) Alert(topic *pb.Topic, stream pb.ReadyNotify_AlertServer) error {
	client := topic.GetClientName()
	appContextID := topic.GetAppContext()
	log.Printf("[ReadyNotify gRPC] Received an Alert subscription request (%s, %s)", client, appContextID)

	// Adding the appContextID entry to the map
	s.mutex.Lock()
	if len(s.alertNotify[appContextID]) == 0 {
		s.alertNotify[appContextID] = make(map[string]pb.ReadyNotify_AlertServer)
	}
	s.alertNotify[appContextID][client] = stream
	s.streamChannel[stream] = make(chan int)
	c := s.streamChannel[stream]
	s.mutex.Unlock()

	// Keep stream open
	for {
		select {
		case <-c:
			log.Printf("[ReadyNotify gRPC] stop channel got triggered for the client = %s\n", client)
			return nil
		}
	}
}

//SendAppContextNotification sends appcontext back to the subscriber if pending
func SendAppContextNotification(appContextID string) error {
	streams := notifServer.alertNotify[appContextID]
	var err error = nil
	for _, stream := range streams {
		err := stream.Send(&pb.Notification{AppContext: appContextID})
		if err != nil {
			logutils.Error("[ReadyNotify gRPC] Notification back to client failed to be sent", logutils.Fields{"err": err})
			// return pkgerrors.New("Notification failed")
		} else {
			logutils.Info("[ReadyNotify gRPC] Notified the subscriber about appContext status changes", logutils.Fields{"appContextID": appContextID})
		}
	}
	return err
}

// Unsubscribe will be called when the subscriber wants to terminate the stream
func (s *readyNotifyServer) Unsubscribe(ctx context.Context, topic *pb.Topic) (*pb.UnsubscribeResponse, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for appContextID, clientStreamMap := range s.alertNotify {
		if appContextID == topic.GetAppContext() {
			stream := clientStreamMap[topic.ClientName]
			s.streamChannel[stream] <- 1

			delete(clientStreamMap, topic.ClientName)
			delete(s.streamChannel, stream)
			// Delete the outer map's appcontextIDs if there is no inner map contents
			if len(clientStreamMap) == 0 {
				delete(s.alertNotify, appContextID)
			}
		}
	}

	return &pb.UnsubscribeResponse{}, nil
}

// NewReadyNotifyServer will create a new readyNotifyServer and destroys the previous one
func NewReadyNotifyServer() *readyNotifyServer {
	s := &readyNotifyServer{
		name:          "readyNotifyServer",
		alertNotify:   make(map[string](map[string]pb.ReadyNotify_AlertServer)),
		streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
	}
	notifServer = s
	return s
}
