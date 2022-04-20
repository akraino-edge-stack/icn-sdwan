// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"sort"

	pkgerrors "github.com/pkg/errors"
)

type NewMockDB struct {
	Items      []map[string]map[string][]byte
	Err        error
	MarshalErr error
}

func (m *NewMockDB) HealthCheck() error {
	return m.Err
}

func createKeyField(key interface{}) (string, error) {

	var n map[string]string
	st, err := json.Marshal(key)
	if err != nil {
		return "", pkgerrors.Errorf("Error Marshalling key: %s", err.Error())
	}
	err = json.Unmarshal([]byte(st), &n)
	if err != nil {
		return "", pkgerrors.Errorf("Error Unmarshalling key to Bson Map: %s", err.Error())
	}
	var keys []string
	for k := range n {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := "{"
	for _, k := range keys {
		s = s + k + ","
	}
	s = s + "}"
	return s, nil
}

func (m *NewMockDB) Insert(table string, key Key, query interface{}, tag string, data interface{}) error {

	i := make(map[string][]byte)
	out, _ := json.Marshal(data)

	// store the tag
	i[tag] = out

	e := make(map[string]map[string][]byte)
	jkey, _ := json.Marshal(key)
	e[string(jkey)] = i

	keymap := map[string]interface{}{}
	json.Unmarshal([]byte(jkey), &keymap)

	// also store all keyvalues so they can be easily retrieved by wildcarding (not just the tags)
	for k, v := range keymap {
		i[k], _ = json.Marshal(v)
	}
	// add the "type" key
	tkey, _ := createKeyField(key)
	i["key"], _ = json.Marshal(tkey)

	// add any extra query key/pair values if supplied
	if query != nil {
		qkey, _ := json.Marshal(query)
		qkeymap := map[string]interface{}{}
		json.Unmarshal([]byte(qkey), &qkeymap)
		for k, v := range qkeymap {
			i[k], _ = json.Marshal(v)
		}
	}

	m.Items = append(m.Items, e)
	return m.Err
}

func (m *NewMockDB) Unmarshal(inp []byte, out interface{}) error {
	err := json.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error Unmarshalling bson data to %T", out)
	}
	return m.MarshalErr
}

func (m *NewMockDB) Find(table string, key Key, tag string) ([][]byte, error) {

	tkey, _ := createKeyField(key)

	newr := make([][]byte, 0)

	// Make match key
	matchkey := make(map[string]string)
	var n map[string]string
	st, _ := json.Marshal(key)
	json.Unmarshal([]byte(st), &n)
	wildmatch := 0
	for k, v := range n {
		if v == "" {
			wildmatch++
		} else {
			matchkey[k] = v
		}
	}
	if wildmatch > 0 {
		matchkey["key"] = tkey
	}

	cnt := 0
	for _, item := range m.Items {
		for _, v := range item {
			// check if matchkey matches this item
			notfound := false
			for mk, mv := range matchkey {
				var iv []byte
				var ok bool
				if iv, ok = v[mk]; !ok {
					notfound = true
					break
				}
				var siv string
				json.Unmarshal(iv, &siv)
				if mv != siv {
					notfound = true
					break
				}
			}
			if notfound {
				break
			}

			// this items key matches - add to the return list if tag is present
			if _, ok := v[tag]; ok {
				newr = append(newr, v[tag])
				cnt++
			}
		}
	}
	if cnt > 0 {
		return newr, m.Err
	} else {
		if m.Err != nil {
			return newr, m.Err
		} else {
			return newr, nil
		}
	}
}

func (m *NewMockDB) Remove(table string, key Key) error {
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

func (m *NewMockDB) RemoveAll(table string, key Key) error {
	return m.Err
}

func (m *NewMockDB) RemoveTag(table string, key Key, tag string) error {
	return m.Err
}
