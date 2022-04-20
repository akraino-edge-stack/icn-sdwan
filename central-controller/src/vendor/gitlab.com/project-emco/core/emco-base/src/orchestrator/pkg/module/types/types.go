// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package types

import (
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"
)

// It implements the interface for managing the ClusterProviders
const MAX_DESCRIPTION_LEN int = 1024
const MAX_USERDATA_LEN int = 4096

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

// Check for valid format Metadata
func IsValidMetadata(metadata Metadata) error {
	errs := validation.IsValidName(metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata name=[%v], errors: %v", metadata.Name, errs)
	}

	errs = validation.IsValidString(metadata.Description, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.Description, errs)
	}

	errs = validation.IsValidString(metadata.UserData1, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.UserData1, errs)
	}

	errs = validation.IsValidString(metadata.UserData2, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.UserData2, errs)
	}

	return nil
}
