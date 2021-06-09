// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package readynotifyserver

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	pb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
)

func Test_readyNotifyServer_Alert(t *testing.T) {
	type fields struct {
		name          string
		alertNotify   map[string]map[string]pb.ReadyNotify_AlertServer
		streamChannel map[pb.ReadyNotify_AlertServer]chan int
		mutex         sync.Mutex
		alertStrm     map[string]pb.ReadyNotify_AlertServer
		appContextId  map[string]string
	}
	type args struct {
		topic  *pb.Topic
		stream pb.ReadyNotify_AlertServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "empty map",
			fields: fields{
				name:          "readyNotifyServer",
				alertNotify:   make(map[string](map[string]pb.ReadyNotify_AlertServer)),
				streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
				alertStrm:     make(map[string]pb.ReadyNotify_AlertServer),
				appContextId:  make(map[string]string),
			},
			args: args{
				topic: &pb.Topic{
					ClientName: "dtc",
					AppContext: "123345",
				},
				stream: nil,
			},
			wantErr: false,
		},
		{
			name: "map has some contents",
			fields: fields{
				name:          "readyNotifyServer",
				alertNotify:   map[string](map[string]pb.ReadyNotify_AlertServer){"123345": {"dtc": nil}},
				streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
				alertStrm:     make(map[string]pb.ReadyNotify_AlertServer),
				appContextId:  make(map[string]string),
			},
			args: args{
				topic: &pb.Topic{
					ClientName: "dtc",
					AppContext: "123345",
				},
				stream: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &readyNotifyServer{
				name:          tt.fields.name,
				alertNotify:   tt.fields.alertNotify,
				streamChannel: tt.fields.streamChannel,
				mutex:         tt.fields.mutex,
			}
			go func() {
				time.Sleep(2 * time.Second)
				s.mutex.Lock()
				s.streamChannel[nil] <- 1
				s.mutex.Unlock()
			}()
			if err := s.Alert(tt.args.topic, tt.args.stream); (err != nil) != tt.wantErr {
				t.Errorf("readyNotifyServer.Alert() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func Test_readyNotifyServer_Unsubscribe(t *testing.T) {
	type fields struct {
		name          string
		alertNotify   map[string]map[string]pb.ReadyNotify_AlertServer
		streamChannel map[pb.ReadyNotify_AlertServer]chan int
		mutex         sync.Mutex
		alertStrm     map[string]pb.ReadyNotify_AlertServer
		appContextId  map[string]string
	}
	type args struct {
		ctx   context.Context
		topic *pb.Topic
	}
	type want struct {
		alertNotify map[string]map[string]pb.ReadyNotify_AlertServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "test1: Delete the contents of the map",
			fields: fields{
				name:          "readyNotifyServer",
				alertNotify:   map[string](map[string]pb.ReadyNotify_AlertServer){"123345": {"dtc": nil}},
				streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
				alertStrm:     make(map[string]pb.ReadyNotify_AlertServer),
				appContextId:  make(map[string]string),
			},
			args: args{
				topic: &pb.Topic{
					ClientName: "dtc",
					AppContext: "123345",
				},
			},
			want: want{
				alertNotify: nil,
			},
			wantErr: false,
		},
		{
			name: "Test2: Delete the contents of the map",
			fields: fields{
				name:          "readyNotifyServer",
				alertNotify:   map[string](map[string]pb.ReadyNotify_AlertServer){"123345": {"dtc": nil}, "245689": {"dcm": nil}},
				streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
				alertStrm:     make(map[string]pb.ReadyNotify_AlertServer),
				appContextId:  make(map[string]string),
			},
			args: args{
				topic: &pb.Topic{
					ClientName: "dtc",
					AppContext: "123345",
				},
			},
			want: want{
				alertNotify: map[string](map[string]pb.ReadyNotify_AlertServer){"245689": {"dcm": nil}},
			},
			wantErr: false,
		},
		{
			name: "Test3: Delete the contents of the map",
			fields: fields{
				name:          "readyNotifyServer",
				alertNotify:   map[string](map[string]pb.ReadyNotify_AlertServer){"123345": {"dtc": nil, "dcm": nil}, "245689": {"dcm": nil}},
				streamChannel: make(map[pb.ReadyNotify_AlertServer]chan int),
				alertStrm:     make(map[string]pb.ReadyNotify_AlertServer),
				appContextId:  make(map[string]string),
			},
			args: args{
				topic: &pb.Topic{
					ClientName: "dtc",
					AppContext: "123345",
				},
			},
			want: want{
				alertNotify: map[string](map[string]pb.ReadyNotify_AlertServer){"123345": {"dcm": nil}, "245689": {"dcm": nil}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			s := &readyNotifyServer{
				name:          tt.fields.name,
				alertNotify:   tt.fields.alertNotify,
				streamChannel: tt.fields.streamChannel,
				mutex:         tt.fields.mutex,
			}

			go func() {
				s.mutex.Lock()
				s.streamChannel[nil] = make(chan int)
				c := s.streamChannel[nil]
				s.mutex.Unlock()
				for {
					time.Sleep(5 * time.Second)
					select {
					case <-c:
						break
					}
				}
			}()
			time.Sleep(2 * time.Second)
			_, err := s.Unsubscribe(tt.args.ctx, tt.args.topic)
			if (err != nil) != tt.wantErr {
				t.Errorf("readyNotifyServer.Unsubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			s.mutex.Lock()
			got := s.alertNotify
			s.mutex.Unlock()
			if !reflect.DeepEqual(len(got), len(tt.want.alertNotify)) {
				t.Errorf("readyNotifyServer.Unsubscribe() = %v, want %v", got, tt.want.alertNotify)
			}

		})
	}
}
