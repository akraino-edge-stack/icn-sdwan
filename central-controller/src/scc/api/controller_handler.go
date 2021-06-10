/*
* Copyright 2020 Intel Corporation, Inc
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
* implied.
* See the License for the specific language governing permissions
* and
* limitations under the License.
 */

package api

import (
	"encoding/json"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/infra/validation"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/manager"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

// ControllerHandler is used to store backend implementations objects
type ControllerHandler struct {
	client manager.ControllerObjectManager
}

// CreateHandler handles creation of the Controller Object entry in the database
func (h ControllerHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	var v module.ControllerObject

	vars := mux.Vars(r)

	v, err = h.client.ParseObject(r.Body)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	validate := validation.GetValidator(h.client.GetStoreMeta())
	isValid, msg := validate.Validate(v)
	if isValid == false {
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}

	// Check resource depedency
	err = manager.GetDBUtils().CheckDep(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check whether the resource is available
	if h.client.IsOperationSupported("GET") {
		vars[h.client.GetResourceName()] = v.GetMetadata().Name
		ret, err = h.client.GetObject(vars)
		if err == nil {
			http.Error(w, "Resource "+v.GetMetadata().Name+" is available already", http.StatusConflict)
			return
		}
	}

	ret, err = h.client.CreateObject(vars, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getsHandler handle GET All operations
func (h ControllerHandler) getsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	// Check resource depedency
	err = manager.GetDBUtils().CheckDep(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ret, err := h.client.GetObjects(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handle GET operations on a particular name
func (h ControllerHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	// Check resource depedency
	err = manager.GetDBUtils().CheckDep(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ret, err := h.client.GetObject(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateHandler handles Update operations
func (h ControllerHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	var v module.ControllerObject

	vars := mux.Vars(r)

	v, err = h.client.ParseObject(r.Body)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	validate := validation.GetValidator(h.client.GetStoreName())
	isValid, msg := validate.Validate(v)
	if isValid == false {
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}

	// Check resource depedency
	err = manager.GetDBUtils().CheckDep(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ret, err = h.client.UpdateObject(vars, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

//deleteHandler handles DELETE operations on a particular record
func (h ControllerHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	// Check resource depedency
	err = manager.GetDBUtils().CheckDep(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check whether sub-resource available
	err = manager.GetDBUtils().CheckOwn(h.client, vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.client.DeleteObject(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
