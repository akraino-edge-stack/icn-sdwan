// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package db

import (
	"encoding/json"
	"reflect"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"

	pkgerrors "github.com/pkg/errors"
)

// DBconn interface used to talk a concrete Database connection
var DBconn Store

// Key is an interface that will be implemented by anypackage
// that wants to use the Store interface. This allows various
// db backends and key types.
type Key interface {
}

// Store is an interface for accessing the database
type Store interface {
	// Returns nil if db health is good
	HealthCheck() error

	// Unmarshal implements any unmarshaling needed for the database
	Unmarshal(inp []byte, out interface{}) error

	// Inserts and Updates a tag with key and also adds query fields if provided
	Insert(coll string, key Key, query interface{}, tag string, data interface{}) error

	// Find the document(s) with key and get the tag values from the document(s)
	Find(coll string, key Key, tag string) ([][]byte, error)

	// Removes the document(s) matching the key if no child reference in collection
	Remove(coll string, key Key) error

	// Remove all the document(s) matching the key
	RemoveAll(coll string, key Key) error

	// Remove the specifiec tag from the document matching the key
	RemoveTag(coll string, key Key, tag string) error
}

// CreateDBClient creates the DB client
func createDBClient(dbType string, dbName string) error {
	var err error
	switch dbType {
	case "mongo":
		// create a mongodb database with orchestrator as the name
		DBconn, err = NewMongoStore(dbName, nil)
	default:
		return pkgerrors.New(dbType + "DB not supported")
	}
	return err
}

// Serialize converts given data into a JSON string
func Serialize(v interface{}) (string, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error serializing "+reflect.TypeOf(v).String())
	}
	return string(out), nil
}

// DeSerialize converts string to a json object specified by type
func DeSerialize(str string, v interface{}) error {
	err := json.Unmarshal([]byte(str), &v)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deSerializing "+str)
	}
	return nil
}

// InitializeDatabaseConnection sets up the connection to the
// configured database to allow the application to talk to it.
func InitializeDatabaseConnection(dbName string) error {
	err := createDBClient(config.GetConfiguration().DatabaseType, dbName)
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = DBconn.HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}
