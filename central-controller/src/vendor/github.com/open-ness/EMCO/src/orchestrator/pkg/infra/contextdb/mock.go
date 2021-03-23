// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextdb

import (
	pkgerrors "github.com/pkg/errors"
)

type MockEtcd struct {
	Items map[string]interface{}
	Err   error
}

func (c *MockEtcd) Put(key string, value interface{}) error {
	if c.Items == nil {
		c.Items = make(map[string]interface{})
	}
	c.Items[key] = value
	return c.Err
}

func (c *MockEtcd) Get(key string, value interface{}) error {
	for kvKey, kvValue := range c.Items {
		if kvKey == key {
			value = kvValue
			return nil
		}
	}
	return pkgerrors.Errorf("Key doesn't exist")
}

func (c *MockEtcd) Delete(key string) error {
	delete(c.Items, key)
	return c.Err
}

func (c *MockEtcd) GetAllKeys(path string) ([]string, error) {
	var keys []string
	for k := range c.Items {
		keys = append(keys, string(k))
	}
	return keys, nil
}

func (e *MockEtcd) HealthCheck() error {
	return nil
}
