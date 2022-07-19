// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package logutils

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
)

// Fields is type that will be used by the calling function
type Fields map[string]interface{}

// Log levels, in order from highest criticality to highest verbosity:
// - panic, fatal, error, warn, info, debug, trace.
// Setting a particular level will show logs of that level and all levels of higher criticality.

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00"})
	// This formatter doesn't support sorting per json keys. It's up to the troubleshooter to
	// pretty-print and/or filter the log. One option to pretty-print while highlighting log levels
	// is to use humanlog (https://github.com/aybabtme/humanlog):
	//
	// From log file:
	//   $ humanlog < log.txt
	// Real time from log file:
	//   $ tail -f log.txt | humanlog
	// Real time from binary:
	//   $ ./rsync 2>&1 | humanlog
	// Real time from docker:
	//   $ docker logs 8e37625ccc9d -f | humanlog

	if strings.EqualFold(config.GetConfiguration().LogLevel, "panic") {
		log.SetLevel(log.PanicLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "fatal") {
		log.SetLevel(log.FatalLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "error") {
		log.SetLevel(log.ErrorLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "warn") {
		log.SetLevel(log.WarnLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "info") {
		log.SetLevel(log.InfoLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "debug") {
		log.SetLevel(log.DebugLevel)
	}
	if strings.EqualFold(config.GetConfiguration().LogLevel, "trace") {
		log.SetLevel(log.TraceLevel)
	}
}

func Panic(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Panic(msg)
	} else {
		log.WithFields(log.Fields(fields)).Panic(msg)
	}
}

func Fatal(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Fatal(msg)
	} else {
		log.WithFields(log.Fields(fields)).Fatal(msg)
	}
}

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

func Debug(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Debug(msg)
	} else {
		log.WithFields(log.Fields(fields)).Debug(msg)
	}
}

func Trace(msg string, fields Fields) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fields != nil {
			fields["SOURCE"] = fmt.Sprintf("file[%s:%d] func[%s]", path.Base(file), line, path.Base(runtime.FuncForPC(pc).Name()))
		}
		log.WithFields(log.Fields(fields)).Trace(msg)
	} else {
		log.WithFields(log.Fields(fields)).Trace(msg)
	}
}

// SetLoglevel .. Set Log level
func SetLoglevel(level log.Level) {
	log.SetLevel(level)
}
