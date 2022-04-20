// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var mockMongoStore *MockMongoStore
var wd string

func init() {
	// make the MongoStore struct hear and then call schema stuff here
	mockMongoStore = &MockMongoStore{
		db: &MockDB{}}

	wd, _ = os.Getwd()
}

var _ = Describe("Read schema file",
	func() {
		Context("when the schema file does not exists", func() {
			It("continue without any error", func() {
				refSchemaFile = "test-schemas/non-exists.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				validate(mockMongoStore.ReadRefSchema(), "")
				clear()
			})
		})

		Context("when the schema file exists but the schema is not valid", func() {
			It("returns schema unmarshal error", func() {
				refSchemaFile = wd + "/test-schemas/invalid-schema.yaml"
				validate(mockMongoStore.ReadRefSchema(), "yaml: line 4: did not find expected key")
				clear()
			})
		})

		Context("when the schema file exists but the schema name is missing", func() {
			It("returns a JSON schema validation error", func() {
				refSchemaFile = wd + "/test-schemas/missing-name.yaml"
				validate(mockMongoStore.ReadRefSchema(), "Invalid Input.")
				clear()
			})
		})

		Context("when the schema file exists but the resources are missing", func() {
			It("returns a JSON schema validation error", func() {
				refSchemaFile = wd + "/test-schemas/missing-resource.yaml"
				validate(mockMongoStore.ReadRefSchema(), "Invalid Input.")
				clear()
			})
		})

		Context("when a valid schema file exists", func() {
			It("continue without any error", func() {
				refSchemaFile = wd + "/test-schemas/emco-base.yaml"
				validate(mockMongoStore.ReadRefSchema(), "")
				clear()
			})

			It("returns a valid schema segment", func() {
				refSchemaFile = wd + "/test-schemas/emco-base.yaml"
				schema, err := readSchema()
				validate(err, "")
				Expect(schema).To(Equal(mockSchema(schema.Name)))
				clear()
			})

			It("returns a valid segmentId", func() {
				refSchemaFile = wd + "/test-schemas/emco-base.yaml"
				rawBytes, _ := ioutil.ReadFile(refSchemaFile)
				segmentId := segmentId(rawBytes)
				Expect(segmentId).To(Equal("d2291e8f0e9fe2fe94b7e440b812448abcfa89a9801d6b78d646553b8ad0a634"))
				clear()
			})
		})
	})

var _ = Describe("Verify referential integrity",
	func() {
		Context("when the schema file does not exists", func() {
			It("create the referential schema map with the schema segments available in the database", func() {
				refSchemaFile = "test-schemas/non-exists.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				refSchemaMapExpected := mockSchemaMap(false)
				validate(mockMongoStore.ReadRefSchema(), "")
				Expect(refSchemaMap).To(Equal(refSchemaMapExpected))
				Expect(len(mockMongoStore.db.Items)).To(Equal(2)) // no new schema segment registered in the database
				clear()
			})
		})

		Context("when the schema file exists", func() {
			It("create the referential schema map with the schema and the schema segments available in the database", func() {
				refSchemaFile = wd + "/test-schemas/new-controller.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				refSchemaMapExpected := mockSchemaMap(true)
				validate(mockMongoStore.ReadRefSchema(), "")
				Expect(refSchemaMap).To(Equal(refSchemaMapExpected))
				Expect(len(mockMongoStore.db.Items)).To(Equal(3)) // register the new schema in the database
				clear()
			})
		},
		)

		Context("when the schema file exists and the controller restarts", func() {
			It("create the referential schema map with the schema segments available in the database", func() {
				refSchemaFile = wd + "/test-schemas/emco-base.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				refSchemaMapExpected := mockSchemaMap(false)
				validate(mockMongoStore.ReadRefSchema(), "")
				Expect(refSchemaMap).To(Equal(refSchemaMapExpected))
				Expect(len(mockMongoStore.db.Items)).To(Equal(2)) // no new schema segment registered in the database
				clear()
			})
		})

		Context("when the dependent schema is missing in the database", func() {
			It("the controller returns an error", func() {
				refSchemaFile = wd + "/test-schemas/missing-parent.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				validate(mockMongoStore.ReadRefSchema(), "Resource schema not found.")
				clear()
			})
		})

		Context("when two controllers define the same resource", func() {
			It("the second controller returns an error", func() {
				refSchemaFile = wd + "/test-schemas/duplicate-resource.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				validate(mockMongoStore.ReadRefSchema(), "Resource already exists.")
				clear()
			})
		})

		Context("when the resource name is invalid", func() {
			It("the controller returns an error", func() {
				refSchemaFile = wd + "/test-schemas/invalid-resource-name.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				validate(mockMongoStore.ReadRefSchema(), "Invalid schema resource name.")
				clear()
			})
		})

		Context("when the resource has a circular dependency on another resource", func() {
			It("the controller returns an error", func() {
				refSchemaFile = wd + "/test-schemas/loop.yaml"
				validate(mockMongoStore.ReadRefSchema(), "Circular schema dependency for resources.")
				clear()
			})
		})

		Context("when two controllers define the same schema name", func() {
			It("the second controller returns an error", func() {
				refSchemaFile = wd + "/test-schemas/duplicate-schema.yaml"
				mockMongoStore.db.Items = mockSchemaSegments()
				validate(mockMongoStore.ReadRefSchema(), "A schema with the name already exists.")
				clear()
			})
		})

	})

func mockSchemaSegments() []map[string]map[string][]byte {
	return []map[string]map[string][]byte{
		{
			DbSchemaKey{
				SegmentId: "d2291e8f0e9fe2fe94b7e440b812448abcfa89a9801d6b78d646553b8ad0a634",
			}.String(): {
				"segment": []byte(
					"{" +
						"\"name\": \"emco-base\"," +
						"\"resources\": [" +
						"{" +
						"\"name\": \"clusterProvider\"," +
						"\"parent\": \"\"," +
						"\"references\": null" +
						"}," +
						"{" +
						"\"name\": \"cluster\"," +
						"\"parent\": \"clusterProvider\"," +
						"\"references\": null" +
						"}," +
						"{" +
						"\"name\": \"clusterLabel\"," +
						"\"parent\": \"cluster\"," +
						"\"references\": null" +
						"}," +
						"{" +
						"\"name\": \"clusterKv\"," +
						"\"parent\": \"cluster\"," +
						"\"references\": null" +
						"}" +
						"]," +
						"\"segmentid\": \"d2291e8f0e9fe2fe94b7e440b812448abcfa89a9801d6b78d646553b8ad0a634\"" +
						"}")},
		},
		{
			DbSchemaKey{
				SegmentId: "e6a83b911fa8f97db7b7b75a1c1ad4c0316cf1e19a652d1326d3b425bfdad9e6",
			}.String(): {
				"segment": []byte(
					"{" +
						"\"name\": \"controller-1\"," +
						"\"resources\": [" +
						"{" +
						"\"name\": \"providerNetwork\"," +
						"\"parent\": \"cluster\"," +
						"\"references\": null" +
						"}," +
						"{" +
						"\"name\": \"network\"," +
						"\"parent\": \"cluster\"," +
						"\"references\": null" +
						"}" +
						"]," +
						"\"segmentid\": \"e6a83b911fa8f97db7b7b75a1c1ad4c0316cf1e19a652d1326d3b425bfdad9e6\"" +
						"}")},
		}}
}

func mockSchemaMap(withSchema bool) map[string]ReferentialSchema {

	refSchemaMap = make(map[string]ReferentialSchema)

	refSchemaMap["clusterProvider"], _ = schemaSegment("clusterProvider", "", "cluster", "", nil, false)
	refSchemaMap["cluster"], _ = schemaSegment("cluster", "clusterProvider", "clusterLabel,clusterKv,providerNetwork,network", "", nil, false)
	refSchemaMap["clusterLabel"], _ = schemaSegment("clusterLabel", "cluster", "", "", nil, false)
	refSchemaMap["clusterKv"], _ = schemaSegment("clusterKv", "cluster", "", "", nil, false)
	refSchemaMap["providerNetwork"], _ = schemaSegment("providerNetwork", "cluster", "", "", nil, false)
	refSchemaMap["network"], _ = schemaSegment("network", "cluster", "", "", nil, false)

	if withSchema { // This is a new schema segment. Include in the ref schema map.
		refSchemaMap["sfcIntent"], _ = schemaSegment("sfcIntent", "", "sfcProviderNetwork,sfcClientSelector,sfcLink", "", nil, false)
		refSchemaMap["sfcClientSelector"], _ = schemaSegment("sfcClientSelector", "sfcIntent", "", "", nil, false)
		refSchemaMap["sfcProviderNetwork"], _ = schemaSegment("sfcProviderNetwork", "sfcIntent", "", "", nil, false)
		refSchemaMap["sfcLink"], _ = schemaSegment("sfcLink", "sfcIntent", "", "", nil, false)
	}

	for resource, schema := range refSchemaMap {
		keyMap, keyId, _, err := createKeyMapAndId(resource)
		if err != nil {
			fmt.Println(err)
			return refSchemaMap
		}
		schema.keyId = keyId
		schema.keyMap = keyMap
		refSchemaMap[resource] = schema
	}

	return refSchemaMap
}

func schemaSegment(resource, parent, children, referencedBy string, references []ReferenceSchema, trimParent bool) (ReferentialSchema, error) {
	mockSchema := ReferentialSchema{
		children:     make(map[string]struct{}), //default
		keyMap:       make(map[string]struct{}), //default
		referencedBy: make(map[string]struct{}), //default
	}

	mockSchema.parent = parent
	mockSchema.trimParent = trimParent
	if len(children) != 0 {
		cn := strings.Split(children, ",")
		for _, c := range cn {
			mockSchema.children[c] = struct{}{}
		}
	}
	if len(referencedBy) != 0 {
		rs := strings.Split(referencedBy, ",")
		for _, r := range rs {
			mockSchema.referencedBy[r] = struct{}{}
		}
	}

	mockSchema.references = references

	return mockSchema, nil
}

func mockSchema(controller string) DbSchema {
	switch controller {
	case "emco-base":
		return DbSchema{
			Name: "emco-base",
			Resources: []ResourceSchema{
				{
					Name:       "clusterProvider",
					Parent:     "",
					References: nil,
				},
				{
					Name:       "cluster",
					Parent:     "clusterProvider",
					References: nil,
				},
				{
					Name:       "clusterLabel",
					Parent:     "cluster",
					References: nil,
				},
				{
					Name:       "clusterKv",
					Parent:     "cluster",
					References: nil,
				},
			},
			SegmentId: "d2291e8f0e9fe2fe94b7e440b812448abcfa89a9801d6b78d646553b8ad0a634"}

	default:
		return DbSchema{}
	}
}

func clear() {
	refSchemaFile = ""
	refSchemaMap = nil
	refKeyMap = nil
	mockMongoStore.db.Items = nil
}

func validate(err error, message string) {
	if len(message) == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(err).To(BeNil())
		return
	}
	Expect(err.Error()).To(ContainSubstring(message))
}
