// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package contextdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

type MockConDb struct {
	Items []sync.Map
	sync.Mutex
	Err   error
}

func (c *MockConDb) Put(key string, value interface{}) error {

	var vg interface{}
	err := c.Get(key, interface{}(&vg))
	if vg != "" {
		c.Delete(key)
	}
//	d := make(map[string][]byte)
	v, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error during json marshal")
	}
//	d[key] = v
	var d sync.Map
	d.Store(key, v)
	c.Lock()
	defer c.Unlock()
	c.Items = append(c.Items, d)
	return c.Err
}
func (c *MockConDb) HealthCheck() error {
	return c.Err
}
func (c *MockConDb) Get(key string, value interface{}) error {
	c.Lock()
	defer c.Unlock()
	for _, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, v := range d {
			if k == key {
				err := json.Unmarshal([]byte(v), value)
				if err != nil {
					fmt.Println("Error during json unmarshal", err, key)
				}
				return c.Err
			}
		}
	}

	value =  nil
	return pkgerrors.Errorf("Key doesn't exist")
}
func (c *MockConDb) GetAllKeys(path string) ([]string, error) {
	c.Lock()
	defer c.Unlock()
	n := 0
	for _, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
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
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
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
	c.Lock()
	defer c.Unlock()
	for i, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
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
	c.Lock()
	defer c.Unlock()
	for i, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
			ok := strings.HasPrefix(k, key)
			if ok {
				c.Items[i] = c.Items[len(c.Items)-1]
				c.Items = c.Items[:len(c.Items)-1]
			}
		}
	}
	return c.Err
}
