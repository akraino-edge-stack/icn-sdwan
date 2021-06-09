// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package logutils

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
	log "github.com/sirupsen/logrus"
)

//Fields is type that will be used by the calling function
type Fields map[string]interface{}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00"})
	if strings.EqualFold(config.GetConfiguration().LogLevel, "warn") {
		log.SetLevel(log.WarnLevel)

	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "info") {
		log.SetLevel(log.InfoLevel)
	}
}

// Error uses the fields provided and logs
func Error(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Error(msg)
	} else {
		log.WithFields(log.Fields(fields)).Error(msg)
	}
}

// Warn uses the fields provided and logs
func Warn(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Warn(msg)
	} else {
		log.WithFields(log.Fields(fields)).Warn(msg)
	}
}

// Info uses the fields provided and logs
func Info(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Info(msg)
	} else {
		log.WithFields(log.Fields(fields)).Info(msg)
	}
}

// SetLoglevel .. Set Log level
func SetLoglevel(level log.Level) {
	log.SetLevel(level)
}
