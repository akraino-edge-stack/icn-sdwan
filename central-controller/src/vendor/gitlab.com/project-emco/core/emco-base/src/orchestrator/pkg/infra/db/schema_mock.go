package db

import (
	pkgerrors "github.com/pkg/errors"
)

type MockMongoStore struct {
	db *MockDB
}

// ReadRefSchema reads the Referential Schema Segment file and creates the refSchemaMap.
func (m *MockMongoStore) ReadRefSchema() error {
	schema, err := readSchema()
	if err != nil {
		return err
	}

	return m.verifyReferentialIntegrity(schema)
}

// verifyReferentialIntegrity verifies the referential integrity of the resources
// defined by the controller(s) schema.
// Wait for controllers to register schema in scenarios where
// multiple controllers start simultaneously.
func (m *MockMongoStore) verifyReferentialIntegrity(serviceSchema DbSchema) error {
	refSchemaMap = nil
	refKeyMap = nil

	waitForSchema, err := m.processSchema(serviceSchema)
	if err != nil {
		return err
	}
	if waitForSchema {
		return pkgerrors.New("Resource schema not found.")
	}

	return nil
}

// processSchema process each schema segment in the db.
func (m *MockMongoStore) processSchema(serviceSchema DbSchema) (bool, error) {
	var (
		emcoRefSchema    DbSchema
		schemaExists     bool
		baseSchemaExists bool
	)

	const baseSchemaName string = "emco-base"

	// Retrieve all the schema segments.
	segments, err := m.db.Find("resources", DbSchemaKey{}, "segment")
	if err != nil {
		return false, err
	}

	if len(segments) == 0 &&
		len(serviceSchema.Resources) == 0 {
		return true, nil
	}

	// Put together a complete schema using the schema segments.
	for _, s := range segments {
		schema := DbSchema{}

		err := m.db.Unmarshal(s, &schema)
		if err != nil {
			return false, err
		}

		if schema.SegmentId == serviceSchema.SegmentId {
			schemaExists = true
		}

		if serviceSchema.Name == schema.Name &&
			serviceSchema.SegmentId != schema.SegmentId {
			return false,
				pkgerrors.New("A schema with the name already exists.")
		}

		if schema.Name == baseSchemaName {
			baseSchemaExists = true
		}

		emcoRefSchema.Resources = append(emcoRefSchema.Resources, schema.Resources...)
	}

	if !baseSchemaExists && serviceSchema.Name != baseSchemaName {
		// Wait for the base schema.
		return true, nil
	}

	if !schemaExists {
		emcoRefSchema.Resources = append(emcoRefSchema.Resources, serviceSchema.Resources...)
	}

	// Create a consolidated referential schemamap.
	err = populateReferentialMap(emcoRefSchema)
	if err != nil {
		return false, err
	}

	// Create a referential keymap.
	waitForSchema, err := populateReferentialKeyMap()
	if err != nil {
		return false, err
	}

	if !schemaExists &&
		serviceSchema.SegmentId != "" {
		// Register the controller schema in the db.
		err := m.db.Insert("resources", DbSchemaKey{SegmentId: serviceSchema.SegmentId}, nil, "segment", serviceSchema)
		if err != nil {
			return false, err
		}
	}

	return waitForSchema, nil
}
