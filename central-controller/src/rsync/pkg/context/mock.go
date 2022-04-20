// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	//"time"
)

// Match stores information about resources applied in clusters
type Match struct {
	// Collects all resources that are deleted
	DeleteMatchList sync.Map
	// Collects all resources that are applied
	ApplyMatchList sync.Map
	// Collects all resources that are currently applied on the cluster
	ResourceList sync.Map
	// Resources committed
	CommitList sync.Map
}

// MatchList to collect resources
var MatchList Match

// MockConnector mocks connector interface
type MockConnector struct {
	sync.Mutex
	Clients *sync.Map
	cid     string
}

func NewProvider(id interface{}) MockConnector {
	MatchList.DeleteMatchList = sync.Map{}
	MatchList.ApplyMatchList = sync.Map{}
	MatchList.ResourceList = sync.Map{}
	MatchList.CommitList = sync.Map{}

	return MockConnector{
		cid: fmt.Sprintf("%v", id),
	}
}

func (c *MockConnector) GetClientProviders(app, cluster, level, namespace string) (ClientProvider, error) {
	c.Lock()
	defer c.Unlock()
	if c.Clients == nil {
		c.Clients = new(sync.Map)
	}
	_, ok := c.Clients.Load(cluster)
	if !ok {
		m := MockClient{cluster: cluster, retryCounter: 0, deletedCounter: 0}
		m.lock = new(sync.Mutex)
		c.Clients.Store(cluster, m)
	}
	m, _ := c.Clients.Load(cluster)
	n := m.(MockClient)
	return &n, nil
}

// MockClient mocks client
type MockClient struct {
	lock           *sync.Mutex
	cluster        string
	retryCounter   int
	deletedCounter int
	applyCounter   int
}

func (m *MockClient) Commit(ctx context.Context, ref interface{}) error {
	str := fmt.Sprintf("%v", ref)
	i, ok := MatchList.CommitList.Load(m.cluster)
	var st string
	if !ok {
		st = string(str)
	} else {
		st = fmt.Sprintf("%v", i) + "," + string(str)
	}
	MatchList.CommitList.Store(m.cluster, st)
	return nil
}
func (m *MockClient) Create(name string, ref interface{}, content []byte) (interface{}, error) {
	return ref, nil
}

func (m *MockClient) TagResource(res []byte, l string) ([]byte, error) {
	return res, nil
}

// Apply Collects resources applied to cluster
func (m *MockClient) Apply(name string, ref interface{}, content []byte) (interface{}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.applyCounter = m.applyCounter + 1
	// Simulate delay
	//time.Sleep(1 * time.Millisecond)
	if len(content) <= 0 {
		return ref, nil
	}
	i, ok := MatchList.ApplyMatchList.Load(m.cluster)
	var str string
	if !ok {
		str = string(content)
	} else {
		str = fmt.Sprintf("%v", i) + "," + string(content)
	}
	MatchList.ApplyMatchList.Store(m.cluster, str)

	i, ok = MatchList.ResourceList.Load(m.cluster)
	if !ok {
		str = string(content)
	} else {
		x := fmt.Sprintf("%v", i)
		if x != "" {
			str = fmt.Sprintf("%v", i) + "," + string(content)
		} else {
			str = string(content)
		}
	}
	MatchList.ResourceList.Store(m.cluster, str)
	if ref != nil {
		str = fmt.Sprintf("%v", ref) + "," + string(content)
	} else {
		str = string(content)
	}
	return str, nil
}

// Delete Collects resources deleted from cluster
func (m *MockClient) Delete(name string, ref interface{}, content []byte) (interface{}, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.deletedCounter = m.deletedCounter + 1
	if len(content) <= 0 {
		return nil, nil
	}
	i, ok := MatchList.DeleteMatchList.Load(m.cluster)
	var str string
	if !ok {
		str = string(content)
	} else {
		str = fmt.Sprintf("%v", i) + "," + string(content)
	}
	MatchList.DeleteMatchList.Store(m.cluster, str)

	// Remove the resource from resourcre list
	i, ok = MatchList.ResourceList.Load(m.cluster)
	if !ok {
		fmt.Println("Deleting resource not applied on cluster", m.cluster)
		return nil, pkgerrors.Errorf("Deleting resource not applied on cluster " + m.cluster)
	} else {
		// Delete it from the string
		a := strings.Split(fmt.Sprintf("%v", i), ",")
		for idx, v := range a {
			if v == string(content) {
				a = append(a[:idx], a[idx+1:]...)
				break
			}
		}
		if len(a) > 0 {
			str = strings.Join(a, ",")
		} else {
			str = ""
		}
		MatchList.ResourceList.Store(m.cluster, str)
	}
	if ref != nil {
		str = fmt.Sprintf("%v", ref) + "," + string(content)
	} else {
		str = string(content)
	}
	return str, nil
}

func (m *MockClient) Get(name string, gvkRes []byte) ([]byte, error) {
	b := []byte("test")
	return b, nil
}
func (m *MockClient) IsReachable() error {
	if m.cluster == "provider1+cluster1" {
		if m.retryCounter < 0 {
			fmt.Println("Counter: ", m.retryCounter)
			m.retryCounter = m.retryCounter + 1
			return pkgerrors.Errorf("Unreachable: " + m.cluster)
		}
	}
	return nil
}

func (m *MockClient) StartClusterWatcher() error {
	return nil
}

func (m *MockClient) ApplyStatusCR(name string, content []byte) error {
	return nil
}

func (m *MockClient) DeleteStatusCR(name string, content []byte) error {
	return nil
}
func (m *MockClient) ApplyConfig(ctx context.Context, config interface{}) error {
	return nil
}

func (m *MockClient) DeleteConfig(ctx context.Context, config interface{}) error {
	return nil
}

func (m *MockClient) CleanClientProvider() error {
	return nil
}

func LoadMap(str string) map[string]string {
	m := make(map[string]string)
	if str == "apply" {
		MatchList.ApplyMatchList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	} else if str == "delete" {
		MatchList.DeleteMatchList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	} else if str == "resource" {
		MatchList.ResourceList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	} else if str == "commit" {
		MatchList.CommitList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	}
	return m
}

func CompareMaps(m, n map[string]string) bool {
	var m1, n1 map[string][]string
	m1 = make(map[string][]string)
	n1 = make(map[string][]string)
	for k, v := range m {
		a := strings.Split(v, ",")
		sort.Strings(a)
		m1[k] = a
	}
	for k, v := range n {
		a := strings.Split(v, ",")
		sort.Strings(a)
		n1[k] = a
	}
	return reflect.DeepEqual(m1, n1)
}

func GetAppContextStatus(cid interface{}, key string) (string, error) {
	//var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return "", err
	}
	hc, err := ac.GetCompositeAppHandle()
	if err != nil {
		return "", err
	}
	dsh, err := ac.GetLevelHandle(hc, key)
	if dsh != nil {
		v, err := ac.GetValue(dsh)
		if err != nil {
			return "", err
		}
		str := fmt.Sprintf("%v", v)
		return str, nil
	}
	return "", err
}

func UpdateAppContextFlag(cid interface{}, key string, b bool) error {
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
	}
	h, err := ac.GetCompositeAppHandle()
	if err != nil {

		return err
	}
	sh, err := ac.GetLevelHandle(h, key)
	if sh == nil {
		_, err = ac.AddLevelValue(h, key, b)
	} else {
		err = ac.UpdateValue(sh, b)
	}
	return err
}
