// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

type MockDB struct {
	Items []map[string]map[string][]byte
	Err   error
}

func (m *MockDB) HealthCheck() error {
	return m.Err
}

func (m *MockDB) Insert(table string, key Key, query interface{}, tag string, data interface{}) error {

	i := make(map[string][]byte)
	out, _ := json.Marshal(data)
	i[tag] = out
	e := make(map[string]map[string][]byte)
	jkey, _ := json.Marshal(key)
	e[string(jkey)] = i
	m.Items = append(m.Items, e)
	return m.Err
}

func (m *MockDB) Unmarshal(inp []byte, out interface{}) error {
	err := json.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling json")
	}
	return nil
}

func (m *MockDB) Find(table string, key Key, tag string) ([][]byte, error) {

	jkey, _ := json.Marshal(key)
	str := (string(jkey))

	var i int
	var r [][]byte
	i = 0
	for _, item := range m.Items {
		for k, _ := range item {
			s := strings.Split(str, "\"\"}")
			if len(s) == 2 {
				ends := strings.TrimPrefix(k, s[0])
				if ends != k && !strings.ContainsAny(ends, ",") && strings.HasSuffix(ends, "}") {
					i++
				}
			} else {
				if str == k {
					break
				}
			}
		}
	}
	if i > 0 {
		r = make([][]byte, i)
	} else {
		r = nil
	}
	i = 0
	for _, item := range m.Items {
		for k, v := range item {
			s := strings.Split(str, "\"\"}")
			if len(s) == 2 {
				ends := strings.TrimPrefix(k, s[0])
				if ends != k && !strings.ContainsAny(ends, ",") && strings.HasSuffix(ends, "}") {
					r[i] = v[tag]
					i++
				}
			} else {
				if str == k {
					return [][]byte{v[tag]}, m.Err
				}
			}
		}
	}
	if i > 0 {
		return r, nil
	} else {
		if m.Err != nil {
			return r, m.Err
		} else {
			return r, pkgerrors.New("Record not found")
		}
	}
}

func (m *MockDB) Remove(table string, key Key) error {
	jkey, _ := json.Marshal(key)
	str := (string(jkey))
	for i, item := range m.Items {
		for k, _ := range item {
			if k == str {
				m.Items[i] = m.Items[len(m.Items)-1]
				m.Items = m.Items[:len(m.Items)-1]
				return m.Err
			}
		}
	}
	return m.Err
}

func (m *MockDB) RemoveAll(table string, key Key) error {
	return m.Err
}

func (m *MockDB) RemoveTag(table string, key Key, tag string) error {
	return m.Err
}
