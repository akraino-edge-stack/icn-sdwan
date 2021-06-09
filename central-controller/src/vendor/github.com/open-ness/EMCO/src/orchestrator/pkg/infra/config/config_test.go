// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package config

import (
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
