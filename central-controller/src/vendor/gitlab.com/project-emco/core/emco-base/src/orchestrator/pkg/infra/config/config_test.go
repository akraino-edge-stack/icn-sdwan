// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package config

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestReadConfigurationFile(t *testing.T) {
	t.Run("Non Existent Configuration File", func(t *testing.T) {
		_, err := readConfigFile("filedoesnotexist.json")
		if err == nil {
			t.Fatal("ReadConfiguationFile: Expected Error, got nil")
		}
	})

	t.Run("Read Configuration File", func(t *testing.T) {
		conf, err := readConfigFile("../../../tests/configs/mock_config.json")
		if err != nil {
			t.Fatal("ReadConfigurationFile: Error reading file: ", err)
		}
		if conf.DatabaseType != "mock_db_test" {
			t.Fatal("ReadConfigurationFile: Incorrect entry read from file")
		}
	})
}

func TestIsValidConfig(t *testing.T) {
	t.Run("Validate Good Configuration File", func(t *testing.T) {
		conf, err := readConfigFile("../../../tests/configs/mock_config.json")
		if err != nil {
			t.Fatal("ReadConfigurationFile: Error reading file: ", err)
		}
		if conf.GrpcCallTimeout != 15 || conf.GrpcConnReadyTime != 1 {
			t.Fatal("Bad values read for some entries")
		}
		isValid := isValidConfig(conf)
		if !isValid {
			t.Fatal("isValidConfig returned failure for good config")
		}
	})
	t.Run("Validate Bad Configuration File", func(t *testing.T) {
		var buf bytes.Buffer
		conf, err := readConfigFile("../../../tests/configs/mock_bad_config.json")
		if err != nil {
			t.Fatal("ReadConfigurationFile: Error reading file: ", err)
		}

		log.SetOutput(&buf) // capture output in a buffer
		defer func() {
			log.SetOutput(os.Stderr)
		}()
		isValid := isValidConfig(conf)
		if isValid {
			t.Fatal("isValidConfig returned success for bad config")
		}
		if !strings.Contains(buf.String(), "GrpcCallTimeout") ||
			!strings.Contains(buf.String(), "GrpcConnReadyTime") {
			t.Fatal("isValidConfig did not report all offending params")
		}
	})
}
