// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"fmt"

	pkgerrors "github.com/pkg/errors"
)

type MockKey struct {
	Key string
}

func (m MockKey) String() string {
	return m.Key
}

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type MockDB struct {
	Store
	Items map[string]map[string][]byte
	Err   error
}

func (m *MockDB) HealthCheck() error {
	return m.Err
}

func (m *MockDB) Insert(table string, key Key, query interface{}, tag string, data interface{}) error {
	return m.Err
}

// MockDB uses simple JSON and not BSON
func (m *MockDB) Unmarshal(inp []byte, out interface{}) error {
	err := json.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling json")
	}
	return nil
}

func (m *MockDB) Find(table string, key Key, tag string) ([][]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	str := fmt.Sprintf("%v", key)
	for k, v := range m.Items {
		if k == str {

			return [][]byte{v[tag]}, nil
		}
	}

	return nil, m.Err
}

func (m *MockDB) Remove(table string, key Key) error {
	return m.Err
}
