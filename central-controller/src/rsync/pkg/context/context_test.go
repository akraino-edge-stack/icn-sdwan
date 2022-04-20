// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context_test

import (
	"testing"
	"time"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

func init() {
	var edb *contextdb.MockConDb
	edb = new(contextdb.MockConDb)
	edb.Err = nil
	contextdb.Db = edb
}

var TestCA CompositeApp = CompositeApp{
	CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
		DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
	AppOrder: []string{"a1", "a2"},
	Apps: map[string]*App{"a1": &App{
		Name: "a1",
		Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
			Name: "provider1+cluster1",
			Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1"},
				"r2": &AppResource{Name: "r2", Data: "a1c1r2"},
			},
			ResOrder: []string{"r1", "r2"}}},
	}, "a2": &App{
		Name: "a2",
		Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
			Name: "provider1+cluster1",
			Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c1r3"},
				"r4": &AppResource{Name: "r4", Data: "a2c1r4"},
			},
			ResOrder: []string{"r3", "r4"}},
			"provider1+cluster2": &Cluster{
				Name: "provider1+cluster2",
				Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c2r3"},
					"r4": &AppResource{Name: "r4", Data: "a2c2r4"},
				},
				ResOrder: []string{"r3", "r4"}}},
	},
	},
}

func TestInstantiateTerminate(t *testing.T) {


	cid, _ := CreateCompApp(TestCA)
	con := NewProvider(cid)

	testCases := []struct {
		label          string
		expectedApply  map[string]string
		expectedDelete map[string]string
		Status         string
		expectedError  error
		event          RsyncEvent
	}{
		{
			expectedApply:  map[string]string{},
			expectedDelete: map[string]string{},
			expectedError:  nil,
			label:          "Read Resources",
			event:          ReadEvent,
		},
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedDelete: map[string]string{},
			expectedError:  nil,
			label:          "Instantiate Resources",
			event:          InstantiateEvent,
		},
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedDelete: map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedError:  nil,
			label:          "Terminate Resources",
			event:          TerminateEvent,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_ = HandleAppContext(cid, nil, testCase.event, &con)
			time.Sleep(2 * time.Second)
			if !CompareMaps(testCase.expectedApply, LoadMap("apply")) {
				t.Error("Apply resources doesn't match", LoadMap("apply"))
			}
			if !CompareMaps(testCase.expectedDelete, LoadMap("delete")) {
				t.Error("Delete resources doesn't match", LoadMap("delete"))
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	testCases := []struct {
		label          string
		original       CompositeApp
		updated        CompositeApp
		expectedApply  map[string]string
		expectedDelete map[string]string
		expectedError  error
	}{
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a1c1r1Updated,a1c1r3"},
			expectedDelete: map[string]string{"provider1+cluster1": "a1c1r2"},
			expectedError:  nil,
			label:          "Update with delete resource and Modify resource",
			original: CompositeApp{
				CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
					DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
				AppOrder: []string{"a1"},
				Apps: map[string]*App{"a1": &App{
					Name: "a1",
					Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
						Name: "provider1+cluster1",
						Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1"},
							"r2": &AppResource{Name: "r2", Data: "a1c1r2"},
						},
						ResOrder: []string{"r1", "r2"}}},
				},
				},
			},
			updated: CompositeApp{
				CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v2", Release: "r1",
					DeploymentIntentGroup: "dig2", Namespace: "default", Level: "0"},
				AppOrder: []string{"a1"},
				Apps: map[string]*App{"a1": &App{
					Name: "a1",
					Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
						Name: "provider1+cluster1",
						Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1Updated"},
							"r3": &AppResource{Name: "r3", Data: "a1c1r3"},
						},
						ResOrder: []string{"r1", "r3"}}},
				},
				},
			},
		},
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a1c1r1Updated,a1c1r3,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedDelete: map[string]string{"provider1+cluster1": "a1c1r2"},
			expectedError:  nil,
			label:          "Update with add new app",
			original: CompositeApp{
				CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
					DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
				AppOrder: []string{"a1"},
				Apps: map[string]*App{"a1": &App{
					Name: "a1",
					Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
						Name: "provider1+cluster1",
						Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1"},
							"r2": &AppResource{Name: "r2", Data: "a1c1r2"},
						},
						ResOrder: []string{"r1", "r2"}}},
				},
				},
			},
			updated: CompositeApp{
				CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v2", Release: "r1",
					DeploymentIntentGroup: "dig2", Namespace: "default", Level: "0"},
				AppOrder: []string{"a1", "a2"},
				Apps: map[string]*App{"a1": &App{
					Name: "a1",
					Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
						Name: "provider1+cluster1",
						Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1Updated"},
							"r3": &AppResource{Name: "r3", Data: "a1c1r3"},
						},
						ResOrder: []string{"r1", "r3"}}},
				},
					"a2": &App{
						Name: "a2",
						Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
							Name: "provider1+cluster1",
							Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c1r3"},
								"r4": &AppResource{Name: "r4", Data: "a2c1r4"},
							},
							ResOrder: []string{"r3", "r4"}},
							"provider1+cluster2": &Cluster{
								Name: "provider1+cluster2",
								Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c2r3"},
									"r4": &AppResource{Name: "r4", Data: "a2c2r4"},
								},
								ResOrder: []string{"r3", "r4"}}},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cid, _ := CreateCompApp(testCase.original)
			ucid, _ := CreateCompApp(testCase.updated)
			con := NewProvider(cid)

			_ = HandleAppContext(cid, nil, InstantiateEvent, &con)
			_ = HandleAppContext(ucid, cid, UpdateEvent, &con)
			time.Sleep(2 * time.Second)

			if !CompareMaps(testCase.expectedApply, LoadMap("apply")) {
				t.Error("Apply resources doesn't match", LoadMap("apply"))
			}
			if !CompareMaps(testCase.expectedDelete, LoadMap("delete")) {
				t.Error("Delete resources doesn't match", LoadMap("delete"))
			}
		})
	}
}

func TestRollbackUpdate(t *testing.T) {

	var original CompositeApp = CompositeApp{
		CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
			DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
		AppOrder: []string{"a1"},
		Apps: map[string]*App{"a1": &App{
			Name: "a1",
			Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
				Name: "provider1+cluster1",
				Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1"},
					"r2": &AppResource{Name: "r2", Data: "a1c1r2"},
				},
				ResOrder: []string{"r1", "r2"}}},
		},
		},
	}

	var updated CompositeApp = CompositeApp{
		CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v2", Release: "r1",
			DeploymentIntentGroup: "dig2", Namespace: "default", Level: "0"},
		AppOrder: []string{"a1", "a2"},
		Apps: map[string]*App{"a1": &App{
			Name: "a1",
			Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
				Name: "provider1+cluster1",
				Resources: map[string]*AppResource{"r1": &AppResource{Name: "r1", Data: "a1c1r1"},
					"r3": &AppResource{Name: "r3", Data: "a1c1r3"},
				},
				ResOrder: []string{"r1", "r3"}}},
		},
			"a2": &App{
				Name: "a2",
				Clusters: map[string]*Cluster{"provider1+cluster1": &Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c1r3"},
						"r4": &AppResource{Name: "r4", Data: "a2c1r4"},
					},
					ResOrder: []string{"r3", "r4"}},
					"provider1+cluster2": &Cluster{
						Name: "provider1+cluster2",
						Resources: map[string]*AppResource{"r3": &AppResource{Name: "r3", Data: "a2c2r3"},
							"r4": &AppResource{Name: "r4", Data: "a2c2r4"},
						},
						ResOrder: []string{"r3", "r4"}}},
			},
		},
	}

	testCases := []struct {
		label                     string
		expectedOriginalResources map[string]string
		expectedUpdatedResources  map[string]string
		expectedError             error
	}{
		{
			expectedOriginalResources: map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2", "provider1+cluster2": ""},
			expectedUpdatedResources:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r3,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedError:             nil,
			label:                     "Test Update with rollback",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cid, _ := CreateCompApp(original)
			ucid, _ := CreateCompApp(updated)
			con := NewProvider(cid)

			_ = HandleAppContext(cid, nil, InstantiateEvent, &con)
			// UPDATE
			_ = HandleAppContext(ucid, cid, UpdateEvent, &con)
			//Update before previous is completed is not supported
			time.Sleep(1 * time.Second)
			if !CompareMaps(testCase.expectedUpdatedResources, LoadMap("resource")) {
				t.Error("Resources doesn't match", LoadMap("resource"))
			}
			// ROLLBACK 1
			_ = HandleAppContext(cid, ucid, UpdateEvent, &con)
			//Update before previous is completed is not supported
			time.Sleep(1 * time.Second)
			if !CompareMaps(testCase.expectedOriginalResources, LoadMap("resource")) {
				t.Error("Resources doesn't match", LoadMap("resource"))
			}
			// ROLLBACK 2
			_ = HandleAppContext(ucid, cid, UpdateEvent, &con)
			time.Sleep(1 * time.Second)
			if !CompareMaps(testCase.expectedUpdatedResources, LoadMap("resource")) {
				t.Error("Resources doesn't match", LoadMap("resource"))
			}
		})
	}
}

func TestStop(t *testing.T) {

	cid, _ := CreateCompApp(TestCA)
	con := NewProvider(cid)

	testCases := []struct {
		label          string
		expectedApply  map[string]string
		expectedDelete map[string]string
		Status         string
		expectedError  error
		stopFlag       bool
	}{
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedDelete: map[string]string{},
			expectedError:  nil,
			label:          "Stop call",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_ = HandleAppContext(cid, nil, InstantiateEvent, &con)
			time.Sleep(2 * time.Second)
			// Set AppContextFlag stop to true
			UpdateAppContextFlag(cid, StopFlagKey, true)
			_ = HandleAppContext(cid, nil, TerminateEvent, &con)
			time.Sleep(1 * time.Second)
			if !CompareMaps(testCase.expectedApply, LoadMap("apply")) {
				t.Error("Apply resources doesn't match", LoadMap("apply"))
			}
			if !CompareMaps(testCase.expectedDelete, LoadMap("delete")) {
				t.Error("Delete resources doesn't match", LoadMap("delete"))
			}
		})
	}
}

func TestInstantiateRestart(t *testing.T) {

	cid, _ := CreateCompApp(TestCA)
	con := NewProvider(cid)

	testCases := []struct {
		label         string
		expectedApply map[string]string
		//expectedDelete map[string]string
		Status        string
		expectedError error
		event         RsyncEvent
	}{
		{
			expectedApply: map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			//expectedDelete: map[string]string{},
			expectedError: nil,
			label:         "Instantiate Resources after restart",
			event:         InstantiateEvent,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_, c := CreateAppContextData(cid)
			_ = c.EnqueueToAppContext(cid, nil, testCase.event)
			_ = RestartAppContext(cid, &con)
			time.Sleep(1 * time.Second)
			if !CompareMaps(testCase.expectedApply, LoadMap("apply")) {
				t.Error("Apply resources doesn't match", LoadMap("apply"))
			}
		})
	}
}

func TestTerminateWithInstantiate(t *testing.T) {

	cid, _ := CreateCompApp(TestCA)
	con := NewProvider(cid)

	testCases := []struct {
		label          string
		expectedApply  map[string]string
		expectedDelete map[string]string
		Status         string
		expectedError  error
		stopFlag       bool
	}{
		{
			expectedApply:  map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedDelete: map[string]string{"provider1+cluster1": "a1c1r1,a1c1r2,a2c1r3,a2c1r4", "provider1+cluster2": "a2c2r3,a2c2r4"},
			expectedError:  nil,
			label:          "Test Terminate With Instantiate Running",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_ = HandleAppContext(cid, nil, InstantiateEvent, &con)
			time.Sleep(1 * time.Millisecond)
			_ = HandleAppContext(cid, nil, TerminateEvent, &con)
			time.Sleep(2 * time.Second)
			if !CompareMaps(testCase.expectedApply, LoadMap("apply")) {
				t.Error("Apply resources doesn't match", LoadMap("apply"))
			}
			if !CompareMaps(testCase.expectedDelete, LoadMap("delete")) {
				t.Error("Delete resources doesn't match", LoadMap("delete"))
			}
		})
	}
}

func TestAppDependency(t *testing.T) {

	var ca CompositeApp = CompositeApp{
		CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
			DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
		AppOrder: []string{"a1", "a2"},
		Apps: map[string]*App{"a1": {
			Name: "a1",
			Clusters: map[string]*Cluster{"provider1+cluster1": {
				Name:      "provider1+cluster1",
				Resources: map[string]*AppResource{"r1": {Name: "r1", Data: "a1c1r1"}},
				ResOrder:  []string{"r1"}}},
			Dependency: map[string]*Criteria{"a2": {OpStatus: "Deployed", Wait: 1}},
		}, "a2": {
			Name: "a2",
			Clusters: map[string]*Cluster{"provider1+cluster1": {
				Name:      "provider1+cluster1",
				Resources: map[string]*AppResource{"r3": {Name: "r3", Data: "a2c1r3"}},
				ResOrder:  []string{"r3"}},
				"provider1+cluster2": {
					Name:      "provider1+cluster2",
					Resources: map[string]*AppResource{"r3": {Name: "r3", Data: "a2c2r3"}},
					ResOrder:  []string{"r3"}}},
		},
		},
	}

	cid, _ := CreateCompApp(ca)
	con := NewProvider(cid)

	testCases := []struct {
		label             string
		expectedResources map[string]string
		Status            string
		expectedError     error
		event             RsyncEvent
	}{
		{
			expectedResources: map[string]string{"provider1+cluster1": "a2c1r3,a1c1r1", "provider1+cluster2": "a2c2r3"},
			expectedError:     nil,
			label:             "Instantiate Resources",
			event:             InstantiateEvent,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_ = HandleAppContext(cid, nil, testCase.event, &con)
			time.Sleep(5 * time.Second)
			if !CompareMaps(testCase.expectedResources, LoadMap("resource")) {
				t.Error("Apply resources doesn't match", LoadMap("resource"), testCase.expectedResources)
			}
		})
	}
}

func setSuccessForAllHooks(cid string, ca CompositeApp) {
	acUtils, _ := utils.NewAppContextReference(cid)
	for _, a := range ca.Apps {
		for _, c := range a.Clusters {
			for _, d := range c.Dependency {
				for _, r := range d {
					acUtils.SetResourceReadyStatus(a.Name, c.Name, r, string(SuccessStatus), true)
				}
			}
		}
	}
}

func TestHooks(t *testing.T) {

	var ca CompositeApp = CompositeApp{
		CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
			DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
		AppOrder: []string{"a1"},
		Apps: map[string]*App{"a1": {
			Name: "a1",
			Clusters: map[string]*Cluster{
				"provider1+cluster1": {
					Name:       "provider1+cluster1",
					Resources:  map[string]*AppResource{"r1+Job": {Name: "r1+Job", Data: "a1c1r1"}, "r2+ConfigMap": {Name: "r2+ConfigMap", Data: "a1c1r2"}, "r3+Pod": {Name: "r3+Pod", Data: "a1c1r3"}},
					Dependency: map[string][]string{"pre-install": {"r1+Job"}, "post-install": {"r2+ConfigMap"}},
					ResOrder:   []string{"r3+Pod"}},
				"provider1+cluster2": {
					Name:       "provider1+cluster2",
					Resources:  map[string]*AppResource{"r1+Job": {Name: "r1+Job", Data: "a1c2r1"}, "r2+ConfigMap": {Name: "r2+ConfigMap", Data: "a1c2r2"}, "r3+Pod": {Name: "r3+Pod", Data: "a1c2r3"}},
					Dependency: map[string][]string{"pre-install": {"r1+Job"}, "post-install": {"r2+ConfigMap"}},
					ResOrder:   []string{"r3+Pod"}}},
		},
		},
	}

	cid, _ := CreateCompApp(ca)
	con := NewProvider(cid)

	testCases := []struct {
		label             string
		expectedResources map[string]string
		Status            string
		expectedError     error
		event             RsyncEvent
	}{
		{
			expectedResources: map[string]string{"provider1+cluster1": "a1c1r1,a1c1r3,a1c1r2", "provider1+cluster2": "a1c2r1,a1c2r3,a1c2r2"},
			expectedError:     nil,
			label:             "Instantiate Resources",
			event:             InstantiateEvent,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_ = HandleAppContext(cid, nil, testCase.event, &con)
			setSuccessForAllHooks(cid, ca)
			time.Sleep(5 * time.Second)
			if !CompareMaps(testCase.expectedResources, LoadMap("resource")) {
				t.Error("Apply resources doesn't match", LoadMap("resource"), testCase.expectedResources)
			}
		})
	}
}

func TestGetAllActiveContext(t *testing.T) {

	cid, _ := CreateCompApp(TestCA)
	_ = NewProvider(cid)

	testCases := []struct {
		label         string
		event         RsyncEvent
		expectedArray []string
		expectedError error
	}{
		{
			label:         "Get all active contexts",
			event:         InstantiateEvent,
			expectedArray: []string{cid},
			expectedError: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_, c := CreateAppContextData(cid)
			_ = c.EnqueueToAppContext(cid, nil, testCase.event)
			UpdateAppContextFlag(cid, StopFlagKey, true)
			cids, _ := GetAllActiveContext()
			time.Sleep(1 * time.Second)
			if len(testCase.expectedArray) == len(cids) {
				for i, v := range testCase.expectedArray {
					if v == cids[i] {
						continue
					} else {
						t.Error("Mismatch in elements", v, cids[i])
					}
				}
			} else {
				t.Error("Mismatch in length of AllActiveContext", len(testCase.expectedArray), len(cids), cids)
			}
		})
	}
}