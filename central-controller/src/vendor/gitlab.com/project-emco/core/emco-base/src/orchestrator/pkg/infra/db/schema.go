// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package db

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gopkg.in/yaml.v3"
)

var (
	// refKeyMap maps the keyId to the name of a resource
	// keyId represents the name of the Keyspace the referential constraint belongs.
	refKeyMap map[string]string

	// refSchemaMap is a map of the schema setup for convenient reference
	// by the DB interface functions to identify and manage dependencies.
	refSchemaMap map[string]ReferentialSchema

	refSchemaFile string = "ref-schemas/v1.yaml"
	schemaLock           = &sync.Mutex{}
)

type ReferenceEntry struct {
	Key   Key
	KeyId string
}

// ReferenceSchema defines the structure of a reference entry in the referential schema.
type ReferenceSchema struct {
	Name       string            `yaml:"name"`
	Type       string            `yaml:"type"`       // can be not present or "map" or "many"
	CommonKey  string            `yaml:"commonKey"`  // optional to disambiguate - part of reference key that is common
	Map        string            `yaml:"map"`        // if Type is "map", this is JSON tag of the Map object
	FixedKv    map[string]string `yaml:"fixedKv"`    // Key/Values of referenced resource that are known at compile time
	FilterKeys []string          `yaml:"filterKeys"` // if "map" type, list of keys to filter (not count as references)
}

// ResourceSchema defines the structure of a data resource in the referential schema.
type ResourceSchema struct {
	Name       string            `yaml:"name"`
	Parent     string            `yaml:"parent"`
	References []ReferenceSchema `yaml:"references"` // if present, overrides default resource id
}

// DbSchema is the top level structure for the referential schema.
type DbSchema struct {
	Name      string           `json:"name"`
	Resources []ResourceSchema `json:"resources"`
	SegmentId string
}

type DbSchemaKey struct {
	SegmentId string
}

// ReferentialSchema defines the structure that will be prepared
// for each element of the referential schema map.
type ReferentialSchema struct {
	parent       string              // Name of the parent resource
	trimParent   bool                // Special case for combined key resource - e.g. compositeApp.compositeAppVersion
	keyId        string              // the keyId string identifier for this resource
	children     map[string]struct{} // list of children
	keyMap       map[string]struct{} // map structure for looking up this item
	references   []ReferenceSchema   // list of references
	referencedBy map[string]struct{} // list of resource that may reference this resource
}

func (key DbSchemaKey) String() string {
	out, err := json.Marshal(key)
	if err != nil {
		return ""
	}
	return string(out)
}

// ReadRefSchema reads the Referential Schema Segment file and creates the refSchemaMap.
func (m *MongoStore) ReadRefSchema() {
	schema, err := readSchema()
	if err != nil {
		return
	}

	m.verifyReferentialIntegrity(schema)
}

// verifyReferentialIntegrity verifies the referential integrity of the resources
// defined by the controller(s) schema.
// Wait for controllers to register schema in scenarios where
// multiple controllers start simultaneously.
func (m *MongoStore) verifyReferentialIntegrity(serviceSchema DbSchema) {
	var (
		backOff       int   = config.GetConfiguration().BackOff
		maxBackOff    int   = config.GetConfiguration().MaxBackOff
		err           error = nil
		waitForSchema bool  = true
	)

	for waitForSchema {
		waitForSchema, err = m.processSchema(serviceSchema)
		if err != nil {
			return
		}
		if !waitForSchema {
			log.Info("DatabaseReferentialSchema: successfully processed the referential schema.",
				log.Fields{})
			break
		}

		log.Info(fmt.Sprintf("DatabaseReferentialSchema: some resources are missing in the schema, retry after %d seconds.", backOff),
			log.Fields{
				"Interval": backOff})
		// Instead of retrying immediately, waits some amount of time between tries.
		time.Sleep(time.Duration(backOff) * time.Second)

		if backOff*2 < maxBackOff {
			backOff *= 2
		} else {
			backOff = maxBackOff
		}
	}
}

// processSchema process each schema segment in the db.
func (m *MongoStore) processSchema(serviceSchema DbSchema) (bool, error) {
	var (
		emcoRefSchema    DbSchema
		schemaExists     bool
		baseSchemaExists bool
	)

	const baseSchemaName string = "emco-base"

	schemaLock.Lock()
	defer schemaLock.Unlock()

	// Retrieve all the schema segments.
	segments, err := m.Find("resources", DbSchemaKey{}, "segment")
	if err != nil {
		log.Error("DatabaseReferentialSchema: failed to retrieve schema segments from db.",
			log.Fields{
				"Error": err})
		return false, err
	}
	if len(segments) == 0 &&
		len(serviceSchema.Resources) == 0 {
		log.Info("DatabaseReferentialSchema: there are no schema(s) registered in the db.",
			log.Fields{})
		// Wait for the schema.
		return true, nil
	}

	// Put together a complete schema using the schema segments.
	for _, s := range segments {
		schema := DbSchema{}
		err := m.Unmarshal(s, &schema)
		if err != nil {
			log.Error("DatabaseReferentialSchema: failed to unmarshal schema segment.",
				log.Fields{
					"Error": err})
			return false, err
		}

		if schema.SegmentId == serviceSchema.SegmentId {
			schemaExists = true
		}

		// Do not register multiple schema segments with the same name.
		if serviceSchema.Name == schema.Name &&
			serviceSchema.SegmentId != schema.SegmentId {
			log.Error("DatabaseReferentialSchema: failed to validate referential schema integrity due to duplicate schema names.",
				log.Fields{
					"Name":             serviceSchema.Name,
					"ExistingSchemaId": schema.SegmentId})
			return false,
				pkgerrors.New("A schema with the name already exists.")
		}

		if schema.Name == baseSchemaName {
			baseSchemaExists = true
		}

		emcoRefSchema.Resources = append(emcoRefSchema.Resources, schema.Resources...)
	}

	if !baseSchemaExists && serviceSchema.Name != baseSchemaName {
		log.Warn("DatabaseReferentialSchema: the emco-base schema is not available.",
			log.Fields{})
		// Wait for the base schema.
		return true, nil
	}

	if !schemaExists {
		emcoRefSchema.Resources = append(emcoRefSchema.Resources, serviceSchema.Resources...)
	}

	// Create a consolidated referential schemamap.
	err = populateReferentialMap(emcoRefSchema)
	if err != nil {
		resetSchema()
		return false, err
	}

	// Create a referential keymap.
	waitForSchema, err := populateReferentialKeyMap()
	if err != nil {
		resetSchema()
		return false, err

	}
	if waitForSchema {
		resetSchema()
	}

	if !schemaExists &&
		serviceSchema.SegmentId != "" {
		// Register the controller schema in the db.
		err := m.Insert("resources", DbSchemaKey{SegmentId: serviceSchema.SegmentId}, nil, "segment", serviceSchema)
		if err != nil {
			log.Error("DatabaseReferentialSchema: failed to insert service schema into the db.",
				log.Fields{
					"Error": err})
			resetSchema()
			return false, err
		}
	}

	return waitForSchema, nil
}

// readSchema reads schema definitions from the given schema file.
func readSchema() (DbSchema, error) {
	var schema DbSchema

	if _, err := os.Stat(refSchemaFile); err != nil {
		if os.IsNotExist(err) {
			log.Warn("DatabaseReferentialSchema: database schema file does not exist.",
				log.Fields{
					"File":  refSchemaFile,
					"Error": err})
			// Continue without a schema
			return schema, nil
		}
		log.Error("DatabaseReferentialSchema: database schema file path error.",
			log.Fields{
				"File":  refSchemaFile,
				"Error": err})
		return schema, err
	}

	rawBytes, err := ioutil.ReadFile(refSchemaFile)
	if err != nil {
		log.Error("DatabaseReferentialSchema: failed to read the database schema file.",
			log.Fields{
				"Error": err,
				"File":  refSchemaFile})
		return schema, err
	}

	err = yaml.Unmarshal(rawBytes, &schema)
	if err != nil {
		log.Error("DatabaseReferentialSchema: failed to unmarshal referential schema.",
			log.Fields{
				"Error": err,
				"File":  refSchemaFile})
		return schema, err
	}

	err = validateSchema(schema)
	if err != nil {
		log.Error("DatabaseReferentialSchema: schema validation failed.",
			log.Fields{
				"Error": err,
				"File":  refSchemaFile})
		return schema, err
	}

	schema.SegmentId = segmentId(rawBytes)

	return schema, nil
}

// populateReferentialMap create a referential schemamap from the schema.
func populateReferentialMap(emcoRefSchema DbSchema) error {
	refSchemaMap = make(map[string]ReferentialSchema)

	for _, resource := range emcoRefSchema.Resources {
		// The name can be of with two elements.
		// eg : compositeApp.compositeAppVersion
		names := strings.Split(resource.Name, ".")
		if len(names) == 0 || len(names) > 2 {
			log.Error("DatabaseReferentialSchema: invalid schema resource name.",
				log.Fields{
					"Resource": resource.Name})
			return pkgerrors.New("Invalid schema resource name.")
		}

		schema := ReferentialSchema{
			children:     make(map[string]struct{}), //default
			keyMap:       make(map[string]struct{}), //default
			referencedBy: make(map[string]struct{}), //default
		}

		if len(names) == 1 { // Resource with only one element - eg: project
			if _, exists := refSchemaMap[names[0]]; !exists {
				schema.parent = resource.Parent
				schema.references = resource.References
				refSchemaMap[names[0]] = schema
				continue
			}
			log.Error("DatabaseReferentialSchema: resource already exists.",
				log.Fields{
					"Resource": names[0]})
			return pkgerrors.New("Resource already exists.")
		}

		// Resource with two elements - eg : compositeApp.compositeAppVersion
		// Handle these two elemets as two separate resource.
		for i, name := range names {
			if _, exists := refSchemaMap[name]; !exists {
				// We take the resource parent as the parent for the first element.
				if i == 0 {
					schema.parent = resource.Parent
					refSchemaMap[name] = schema
					continue
				}
				// We take the name of the first element as the parent of the second element.
				schema.parent = names[0]
				schema.trimParent = true
				schema.references = resource.References
				refSchemaMap[name] = schema
				continue
			}
			log.Error("DatabaseReferentialSchema: resource already exists.",
				log.Fields{
					"Resource": name})
			return pkgerrors.New("Resource already exists.")
		}
	}

	return nil
}

// populateReferentialKeyMap create a referential keymap from the referential schemamap.
func populateReferentialKeyMap() (bool, error) {
	refKeyMap = make(map[string]string)

	for resource, schema := range refSchemaMap {
		if schema.parent != "" {
			p, exists := refSchemaMap[schema.parent]
			if !exists {
				log.Warn("DatabaseReferentialSchema: parent resource is missing in the referential schema map.",
					log.Fields{
						"Resource": resource,
						"Parent":   schema.parent})
				// Wait for the missing schema.
				return true, nil
			}
			// Add to parents child list.
			p.children[resource] = struct{}{}
			refSchemaMap[schema.parent] = p
		}

		// Check if "referenced by" and update referencedBy list of references.
		for _, reference := range schema.references {
			r, exists := refSchemaMap[reference.Name]
			if !exists {
				log.Warn("DatabaseReferentialSchema: referenced resource is missing in the referential schema map.",
					log.Fields{
						"Resource":           resource,
						"ReferencedResource": reference.Name})
				// Wait for the missing schema.
				return true, nil
			}
			// Add to referencedBy list
			r.referencedBy[resource] = struct{}{}
			refSchemaMap[reference.Name] = r
		}

		keyMap, keyId, waitForSchema, err := createKeyMapAndId(resource)
		if err != nil {
			return false, err
		}
		if waitForSchema {
			return waitForSchema, err
		}

		schema.keyMap = keyMap
		schema.keyId = keyId
		refSchemaMap[resource] = schema
		refKeyMap[keyId] = resource
	}

	return false, nil
}

// createKeyMapAndId returns key structure (map) and keyId for a given resource.
// KeyId indicates the referential integrity keys for the resource.
func createKeyMapAndId(resource string) (map[string]struct{}, string, bool, error) {
	schema, exists := refSchemaMap[resource]
	if !exists {
		log.Warn("DatabaseReferentialSchema: resource is missing in the referential schema map.",
			log.Fields{
				"Resource": resource})
		// Wait for the missing schema.
		return nil, "", true, nil
	}

	var keys []string
	key := resource
	keyMap := make(map[string]struct{})

	// Generate the keyMap.
	for {
		if _, exists = keyMap[key]; exists {
			log.Error("DatabaseReferentialSchema: circular schema dependency for resource.",
				log.Fields{
					"Resource": key})
			return nil, "", false, pkgerrors.New("Circular schema dependency for resources.")
		}

		keys = append(keys, key)
		keyMap[key] = struct{}{}

		key = schema.parent
		if len(key) == 0 {
			break
		}

		if schema, exists = refSchemaMap[key]; !exists {
			log.Warn("DatabaseReferentialSchema: parent resource is missing in the referential schema map.",
				log.Fields{
					"Resource": resource,
					"Parent":   key})
			// Wait for the missing schema.
			return nil, "", true, nil
		}
	}

	sort.Strings(keys)

	// Generate the keyId
	keyId := fmt.Sprintf("{%s}", strings.Join(keys, ","))

	return keyMap, keyId, false, nil
}

// segmentId returns a unique id for the segment.
func segmentId(p []byte) string {
	h := sha256.New()
	h.Write(p)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// validateSchema validate the provided data with the schema definition
func validateSchema(schema DbSchema) error {
	var err error

	// Schema Name is required.
	if err = match(schema.Name, true); err != nil {
		return err
	}

	// Resource(s) are required.
	if len(schema.Resources) == 0 {
		log.Error("DatabaseReferentialSchema: invalid Input:: Invalid type. Expected: array, given: null",
			log.Fields{
				"Schema": schema.Name,
				"Field":  "Resources"})
		return pkgerrors.New("Invalid Input.")
	}

	for _, res := range schema.Resources {
		// Resource Name is required.
		if err = match(res.Name, true); err != nil {
			return err
		}

		if err = match(res.Parent, false); err != nil {
			return err
		}

		for _, ref := range res.References {
			// Reference Name is required.
			if err = match(ref.Name, true); err != nil {
				return err
			}

			if err = match(ref.Map, false); err != nil {
				return err
			}

			if err = match(ref.CommonKey, false); err != nil {
				return err
			}

			// Allowed Types are 'map', 'many', if present
			if len(ref.Type) != 0 &&
				ref.Type != "map" &&
				ref.Type != "many" {
				log.Error("DatabaseReferentialSchema: invalid Input:: Invalid value. Expected: map or many",
					log.Fields{
						"Schema": schema.Name,
						"Field":  "Reference Type",
						"Value":  ref.Type})
				return pkgerrors.New("Invalid Input.")
			}

			for _, kv := range ref.FixedKv {
				if err = match(kv, false); err != nil {
					return err
				}
			}

			for _, k := range ref.FilterKeys {
				if err = match(k, false); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// match matches the given string with a pattern
func match(s string, isRequired bool) error {
	r, err := regexp.Compile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")
	if err != nil {
		return err
	}

	m := r.MatchString(s)
	if (isRequired && len(s) == 0) ||
		(len(s) != 0 && !m) {
		log.Error("DatabaseReferentialSchema: invalid Input:: Does not match pattern '^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$'",
			log.Fields{
				"Value": s})
		return pkgerrors.New("Invalid Input.")
	}

	return nil
}

// reset clears the ref map
func resetSchema() {
	refSchemaMap = nil
	refKeyMap = nil
}
