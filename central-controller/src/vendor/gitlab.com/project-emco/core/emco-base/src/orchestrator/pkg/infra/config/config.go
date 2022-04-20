// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package config

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strings"
)

// Configuration loads up all the values that are used to configure
// backend implementations
type Configuration struct {
	CAFile                 string `json:"ca-file"`
	ServerCert             string `json:"server-cert"`
	ServerKey              string `json:"server-key"`
	Password               string `json:"password"`
	DatabaseIP             string `json:"database-ip"`
	DatabaseType           string `json:"database-type"`
	PluginDir              string `json:"plugin-dir"`
	EtcdIP                 string `json:"etcd-ip"`
	EtcdCert               string `json:"etcd-cert"`
	EtcdKey                string `json:"etcd-key"`
	EtcdCAFile             string `json:"etcd-ca-file"`
	GrpcServerCert         string `json:"grpc-server-cert"`
	GrpcServerKey          string `json:"grpc-server-key"`
	GrpcCAFile             string `json:"grpc-ca-file"`
	GrpcEnableTLS          string `json:"grpc-enable-tls"`
	GrpcServerNameOverride string `json:"grpc-server-name-override"`
	ServicePort            string `json:"service-port"`
	KubernetesLabelName    string `json:"kubernetes-label-name"`
	LogLevel               string `json:"log-level"`
	MaxRetries             string `json:"max-retries"`
	BackOff                int    `json:"db-schema-backoff"`
	MaxBackOff             int    `json:"db-schema-max-backoff"`

	// EMCO-internal communication
	//    wait time for a grpc connection to become ready, in milliseconds
	GrpcConnReadyTime int `json:"grpc-conn-ready-time"`
	//    wait time to declare a grpc conn as failed, in milliseconds
	GrpcConnTimeout int `json:"grpc-conn-timeout"`
	//    RPC call timeout, in milliseconds
	//    gRPC call deadline may vary across controllers. We could have each
	//    controller register with orch with a timeout value, in the future.
	//    For now, we use a fixed timeout for all.
	GrpcCallTimeout int `json:"grpc-call-timeout"`

	// TODO: EMCO-K8s communication: Create similar time/timeout params
}

// Config is the structure that stores the configuration
var gConfig *Configuration

// readConfigFile reads the specified smsConfig file to setup some env variables
func readConfigFile(file string) (*Configuration, error) {
	f, err := os.Open(file)
	if err != nil {
		return defaultConfiguration(), err
	}
	defer f.Close()

	// Setup some defaults here
	// If the json file has values in it, the defaults will be overwritten
	conf := defaultConfiguration()

	// Read the configuration from json file
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func defaultConfiguration() *Configuration {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Error getting cwd. Using .")
		cwd = "."
	}

	return &Configuration{
		CAFile:                 "ca.cert",
		ServerCert:             "server.cert",
		ServerKey:              "server.key",
		Password:               "",
		DatabaseIP:             "127.0.0.1",
		DatabaseType:           "mongo",
		PluginDir:              cwd,
		EtcdIP:                 "127.0.0.1",
		EtcdCert:               "",
		EtcdKey:                "",
		EtcdCAFile:             "",
		GrpcServerCert:         "",
		GrpcServerKey:          "",
		GrpcCAFile:             "",
		GrpcEnableTLS:          "disable",
		GrpcServerNameOverride: "",
		ServicePort:            "9015",
		KubernetesLabelName:    "orchestrator.io/rb-instance-id",
		LogLevel:               "warn", // default log-level of all modules
		MaxRetries:             "",     // rsync
		BackOff:                5,      // default backoff time interval for ref schema
		MaxBackOff:             60,     // max backoff time interval for ref schema
		GrpcConnReadyTime:      1000,   // 1 second in milliseconds
		GrpcConnTimeout:        1000,   // 1 second
		GrpcCallTimeout:        10000,  // 10 seconds
	}
}

func isValidConfig(cfg *Configuration) bool {
	valid := true
	members := reflect.ValueOf(cfg).Elem()

	// If a config param has "Time" in its name, and is type int,
	// ensure its value is positive.
	for i := 0; i < members.NumField(); i++ {
		varName := members.Type().Field(i).Name
		varValue := members.Field(i).Interface()
		if strings.Contains(varName, "Time") {
			intValue, ok := varValue.(int)
			if ok && intValue <= 0 {
				log.Printf("%s must be positive, not %d.\n",
					varName, intValue)
				valid = false
			}
		}
	}
	return valid
}

// GetConfiguration returns the configuration for the app.
// It will try to load it if it is not already loaded.
func GetConfiguration() *Configuration {
	if gConfig == nil {
		conf, err := readConfigFile("config.json")
		if err != nil {
			log.Println("Error loading config file: ", err)
			log.Println("Using defaults...")
		}
		gConfig = conf

		if !isValidConfig(gConfig) {
			log.Fatalln("Bad data in config. Exiting.")
			return nil
		}
	}

	return gConfig
}

// SetConfigValue sets a value in the configuration
// This is mostly used to customize the application and
// should be used carefully.
func SetConfigValue(key string, value string) *Configuration {
	c := GetConfiguration()
	if value == "" || key == "" {
		return c
	}

	v := reflect.ValueOf(c).Elem()
	if v.Kind() == reflect.Struct {
		f := v.FieldByName(key)
		if f.IsValid() {
			if f.CanSet() {
				if f.Kind() == reflect.String {
					f.SetString(value)
				}
			}
		}
	}
	return c
}
