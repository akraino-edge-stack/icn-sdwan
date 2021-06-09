// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package contextdb

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MockConDb struct {
	Items []map[string][]byte
	Err   error
}

func (c *MockConDb) Put(key string, value interface{}) error {

	var vg interface{}
	err := c.Get(key, interface{}(&vg))
	if vg != "" {
		c.Delete(key)
	}
	d := make(map[string][]byte)
	v, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error during json marshal")
	}
	d[key] = v
	c.Items = append(c.Items, d)
	return c.Err
}
func (c *MockConDb) HealthCheck() error {
	return c.Err
}
func (c *MockConDb) Get(key string, value interface{}) error {
	for _, item := range c.Items {
		for k, v := range item {
			if k == key {
				err := json.Unmarshal([]byte(v), value)
				if err != nil {
					fmt.Println("Error during json unmarshal")
				}
				return c.Err
			}
		}
	}
	value = nil
	return c.Err
}
func (c *MockConDb) GetAllKeys(path string) ([]string, error) {
	n := 0
	for _, item := range c.Items {
		for k, _ := range item {
			ok := strings.HasPrefix(k, path)
			if ok {
				n++
			}
		}
	}
	if n == 0 {
		return nil, c.Err
	}

	retk := make([]string, n)

	i := 0
	for _, item := range c.Items {
		for k, _ := range item {
			ok := strings.HasPrefix(k, path)
			if ok {
				retk[i] = k
				i++
			}
		}
	}
	return retk, c.Err
}
func (c *MockConDb) Delete(key string) error {
	for i, item := range c.Items {
		for k, _ := range item {
			if k == key {
				c.Items[i] = c.Items[len(c.Items)-1]
				c.Items = c.Items[:len(c.Items)-1]
				return c.Err
			}
		}
	}
	return c.Err
}
func (c *MockConDb) DeleteAll(key string) error {
	for i, item := range c.Items {
		for k, _ := range item {
			ok := strings.HasPrefix(k, key)
			if ok {
				c.Items[i] = c.Items[len(c.Items)-1]
				c.Items = c.Items[:len(c.Items)-1]
			}
		}
	}
	return c.Err
}
