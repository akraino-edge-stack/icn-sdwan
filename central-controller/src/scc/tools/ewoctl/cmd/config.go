// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"
	"os"
	"strconv"
)

// Configurations exported
type EwoConfigurations struct {
	Ingress      ControllerConfigurations
	Ewo          ControllerConfigurations
}

// ControllerConfigurations exported
type ControllerConfigurations struct {
	Port int
	Host string
}

const baseUrl string = "scc/v1"
const urlPrefix string = "http://"

var Configurations EwoConfigurations

// SetDefaultConfiguration default configuration if t
func SetDefaultConfiguration() {
	Configurations.Ewo.Host = "localhost"
	Configurations.Ewo.Port = 9015
}

// GetIngressURL Url for Ingress
func GetIngressURL() string {
	if Configurations.Ingress.Host == "" || Configurations.Ingress.Port == 0 {
		return ""
	}
	return urlPrefix + Configurations.Ingress.Host + ":" + strconv.Itoa(Configurations.Ingress.Port) + "/" + baseUrl
}

// GetEwoURL Url for Edge Wan Overlay Controller
func GetEwoURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Ewo.Host == "" || Configurations.Ewo.Port == 0 {
		fmt.Println("Fatal: No Ewo Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.Ewo.Host + ":" + strconv.Itoa(Configurations.Ewo.Port) + "/" + baseUrl
}