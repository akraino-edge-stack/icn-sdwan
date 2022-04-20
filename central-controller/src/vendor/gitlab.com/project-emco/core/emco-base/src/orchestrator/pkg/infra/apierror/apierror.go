// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package apierror

import (
	"net/http"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type APIError struct {
	ID      string
	Message string
	Status  int
}

var dbErrors = []APIError{
	{ID: "db Find error", Message: "Error finding referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Remove error", Message: "Error removing referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Remove resource not found", Message: "The requested resource not found", Status: http.StatusNotFound},
	{ID: "db Remove parent child constraint", Message: "Cannot delete parent without deleting child references first", Status: http.StatusConflict},
	{ID: "db Remove referential constraint", Message: "Cannot delete without deleting or updating referencing resources first", Status: http.StatusConflict},
	{ID: "db Insert error", Message: "Error adding or updating referencing resources", Status: http.StatusInternalServerError},
	{ID: "db Insert parent resource not found", Message: "Cannot perform requested operation. Parent resource not found", Status: http.StatusConflict},
	{ID: "db Insert referential schema missing", Message: "Cannot perform requested operation. The requested resource is not defined in the referential schema", Status: http.StatusConflict},
}

// shared list the errors a controller can get from a dependent controller
// for example, DTC calls the orchestrator to check the status of a deployment intent group
// and the orchestrator returns `DeploymentIntentGroup not found`
var shared = []APIError{
	{ID: "App not found", Message: "app not found", Status: http.StatusNotFound},
	{ID: "AppDependency not found", Message: "appDependency not found", Status: http.StatusNotFound},
	{ID: "AppIntent not found", Message: "appIntent not found", Status: http.StatusNotFound},
	{ID: "AppProfile not found", Message: "appProfile not found", Status: http.StatusNotFound},
	{ID: "CompositeApp not found", Message: "compositeApp not found", Status: http.StatusNotFound},
	{ID: "CompositeProfile not found", Message: "compositeProfile not found", Status: http.StatusNotFound},
	{ID: "Controller not found", Message: "controller not found", Status: http.StatusNotFound},
	{ID: "Cluster not found", Message: "cluster not found", Status: http.StatusNotFound},
	{ID: "DeploymentIntentGroup not found", Message: "deploymentIntentGroup not found", Status: http.StatusNotFound},
	{ID: "GenericPlacementIntent not found", Message: "genericPlacementIntent not found", Status: http.StatusNotFound},
	{ID: "Intent not found", Message: "intent not found", Status: http.StatusNotFound},
	{ID: "LogicalCloud not found", Message: "logicalCloud not found", Status: http.StatusNotFound},
	{ID: "Project not found", Message: "project not found", Status: http.StatusNotFound},
}

// HandleErrors handles api resources add/update/create errors
// Returns APIError with the ID, message and the http status based on the error
func HandleErrors(params map[string]string, err error, mod interface{}, apiErr []APIError) APIError {
	log.Error("Error :: ", log.Fields{"Parameters": params, "Error": err, "Module": mod})

	// db errors
	for _, e := range dbErrors {
		if strings.Contains(err.Error(), e.ID) {
			return e
		}
	}

	// api specific errors
	for _, e := range apiErr {
		if strings.Contains(err.Error(), e.ID) {
			return e
		}
	}

	// conditional errors
	for _, e := range shared {
		if strings.Contains(err.Error(), e.ID) {
			return e
		}
	}

	// Default
	return APIError{ID: "Internal server error", Message: "The server encountered an internal error and was unable to complete your request.", Status: http.StatusInternalServerError}
}

// HandleLogicalCloudErrors handles logical cloud errors
// Returns APIError with the ID, message and the http status based on the error
func HandleLogicalCloudErrors(params map[string]string, err error, lcErrors []APIError) APIError {
	for _, e := range lcErrors {
		if strings.Contains(err.Error(), e.ID) {
			log.Error("Logical cloud error :: ", log.Fields{"Parameters": params, "Error": err})
			return e
		}
	}
	return APIError{}
}
