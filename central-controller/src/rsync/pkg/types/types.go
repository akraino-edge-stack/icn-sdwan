// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package types

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
)

const (
	CurrentStateKey         string = "rsync/state/CurrentState"
	DesiredStateKey         string = "rsync/state/DesiredState"
	PendingTerminateFlagKey string = "rsync/state/PendingTerminateFlag"
	AppContextEventQueueKey string = "rsync/AppContextEventQueue"
	StatusKey               string = "status"
	StopFlagKey             string = "stopflag"
	StatusAppContextIDKey   string = "statusappctxid"
)

// RsyncEvent is event Rsync handles
type RsyncEvent string

// Rsync Event types
const (
	InstantiateEvent     RsyncEvent = "Instantiate"
	TerminateEvent       RsyncEvent = "Terminate"
	ReadEvent            RsyncEvent = "Read"
	AddChildContextEvent RsyncEvent = "AddChildContext"
	UpdateEvent          RsyncEvent = "Update"
	// This is an internal event
	UpdateDeleteEvent RsyncEvent = "UpdateDelete"
)

// RsyncOperation is operation Rsync handles
type RsyncOperation int

// Rsync Operations
const (
	OpApply RsyncOperation = iota
	OpDelete
	OpRead
	OpCreate
)

// ResourceStatusType defines types of resource statuses
type ResourceStatusType string

// ResourceStatusType
const (
	ReadyStatus   ResourceStatusType = "resready"
	SuccessStatus ResourceStatusType = "ressuccess"
)

func (d RsyncOperation) String() string {
	return [...]string{"Apply", "Delete", "Read"}[d]
}

// StateChange represents a state change rsync handles
type StateChange struct {
	// Name of the event
	Event RsyncEvent
	// List of states that can handle this event
	SState []appcontext.StatusValue
	// Dst state if the state transition was successful
	DState appcontext.StatusValue
	// Current state if the state transition was successful
	CState appcontext.StatusValue
	// Error state if the state transition was unsuccessful
	ErrState appcontext.StatusValue
}

// StateChanges represent State Machine for the AppContext
var StateChanges = map[RsyncEvent]StateChange{

	InstantiateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
	TerminateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.TerminateFailed,
			appcontext.AppContextStatusEnum.Terminating,
			appcontext.AppContextStatusEnum.Updated,
			appcontext.AppContextStatusEnum.UpdateFailed,
			appcontext.AppContextStatusEnum.Updating,
		},
		DState:   appcontext.AppContextStatusEnum.Terminated,
		CState:   appcontext.AppContextStatusEnum.Terminating,
		ErrState: appcontext.AppContextStatusEnum.TerminateFailed,
	},
	UpdateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Updated},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
	UpdateDeleteEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Instantiated},
		DState:   appcontext.AppContextStatusEnum.Updated,
		CState:   appcontext.AppContextStatusEnum.Updating,
		ErrState: appcontext.AppContextStatusEnum.UpdateFailed,
	},
	ReadEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
}

// Resource Dependency Structures
// RsyncEvent is event Rsync handles
type OpStatus string

// OpStatus types
const (
	OpStatusDeployed OpStatus = "Deployed"
	OpStatusReady    OpStatus = "Ready"
	OpStatusDeleted  OpStatus = "Deleted"
)

type Resource struct {
	App string                  `json:"app,omitempty"`
	Res string                  `json:"name,omitempty"`
	GVK schema.GroupVersionKind `json:"gvk,omitempty"`
}

// Criteria for Resource dependency
type Criteria struct {
	// Ready or deployed
	OpStatus OpStatus `json:"opstatus,omitempty"`
	// Wait time in seconds
	Wait int `json:"wait,omitempty"`
}

type AppCriteria struct {
	App string `json:"app"`
	// Ready or deployed
	OpStatus OpStatus `json:"opstatus,omitempty"`
	// Wait time in seconds
	Wait int `json:"wait,omitempty"`
}

// Dependency Structures
type Dependency struct {
	Resource Resource `json:"resource,omitempty"`
	Criteria Criteria `json:"criteria,omitempty"`
}

// ResourceDependency structure
type ResourceDependency struct {
	Resource Resource     `json:"resource,omitempty"`
	Dep      []Dependency `json:"dependency,omitempty"`
}

// CompositeApp Structures
type CompositeApp struct {
	Name         string                      `json:"name,omitempty"`
	CompMetadata appcontext.CompositeAppMeta `json:"compmetadat,omitempty"`
	AppOrder     []string                    `json:"appOrder,omitempty"`
	Apps         map[string]*App             `json:"apps,omitempty"`
}

// AppResource represents a resource
type AppResource struct {
	Name string      `json:"name,omitempty"`
	Data interface{} `json:"data,omitempty"`
	// Needed to suport updates
	Skip bool `json:"bool,omitempty"`
}

// Cluster is a cluster within an App
type Cluster struct {
	Name       string                  `json:"name,omitempty"`
	ResOrder   []string                `json:"reorder,omitempty"`
	Resources  map[string]*AppResource `json:"resources,omitempty"`
	Dependency map[string][]string     `json:"resdependency,omitempty"`
	// Needed to suport updates
	Skip bool `json:"bool,omitempty"`
}

// App is an app within a composite app
type App struct {
	Name       string               `json:"name,omitempty"`
	Clusters   map[string]*Cluster  `json:"clusters,omitempty"`
	Dependency map[string]*Criteria `json:"dependency,omitempty"`
	// Needed to suport updates
	Skip bool `json:"bool,omitempty"`
}

// ResourceProvider is interface for working with the resources
type ResourceProvider interface {
	Create(name string, ref interface{}, content []byte) (interface{}, error)
	Apply(name string, ref interface{}, content []byte) (interface{}, error)
	Delete(name string, ref interface{}, content []byte) (interface{}, error)
	Get(name string, gvkRes []byte) ([]byte, error)
	Commit(ctx context.Context, ref interface{}) error
	IsReachable() error
	TagResource([]byte, string) ([]byte, error)
}

type StatusProvider interface {
	StartClusterWatcher() error
	ApplyStatusCR(name string, content []byte) error
	DeleteStatusCR(name string, content []byte) error
}

type ReferenceProvider interface {
	ApplyConfig(ctx context.Context, config interface{}) error
	DeleteConfig(ctx context.Context, config interface{}) error
}

// Client Provider provides functionality to interface with the cluster
type ClientProvider interface {
	ResourceProvider
	StatusProvider
	ReferenceProvider
	CleanClientProvider() error
}

// Connection is interface for connection
type Connector interface {
	GetClientProviders(app, cluster, level, namespace string) (ClientProvider, error)
}

// AppContextQueueElement element in per AppContext Queue
type AppContextQueueElement struct {
	Event RsyncEvent `json:"event"`
	// Only valid in case of update events
	UCID string `json:"uCID,omitempty"`
	// Status - Pending, Done, Error, skip
	Status string `json:"status"`
}

// AppContextQueue per AppContext queue
type AppContextQueue struct {
	AcQueue []AppContextQueueElement
}
