/*
Copyright (c) 2021 The cirocket Authors (Neil Hemming)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package stdlog implements a loggee wrapper around the standard log package
package stdlog

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nehemming/cirocket/pkg/loggee"
)

type (
	stdLog struct {
		l loggee.Level
	}
)

// New creates a new logger using the standard golang 'log' package.
func New() loggee.Logger {
	return &stdLog{}
}

// Activity is a nested activity.
func (std *stdLog) Activity(ctx context.Context, fn loggee.ActivityFunc) error {
	return fn(ctx)
}

// SetLevel sets the logging level.
func (std *stdLog) SetLevel(l loggee.Level) {
	std.l = l
}

// WithFields adds fields to a log Entry.
func (std *stdLog) WithFields(f loggee.Fielder) loggee.Entry {
	return newEntry().WithFields(f)
}

// WithField adds a field to an Entry.
func (std *stdLog) WithField(key string, value interface{}) loggee.Entry {
	return newEntry().WithField(key, value)
}

// WithDuration adds a time t an Entry.
func (std *stdLog) WithDuration(d time.Duration) loggee.Entry {
	return newEntry().WithDuration(d)
}

// WithError adds an error to the Entry.
func (std *stdLog) WithError(e error) loggee.Entry {
	return newEntry().WithError(e)
}

// Debug writes a debug message.
func (std *stdLog) Debug(msg string) {
	if std.l == loggee.DebugLevel {
		log.Println("debug", msg)
	}
}

// Info writes a info message.
func (std *stdLog) Info(msg string) {
	if std.l <= loggee.InfoLevel {
		log.Println("info", msg)
	}
}

// Warn writes a warning message.
func (std *stdLog) Warn(msg string) {
	if std.l <= loggee.WarnLevel {
		log.Println("warn", msg)
	}
}

// Error writes an error message.
func (std *stdLog) Error(msg string) {
	if std.l <= loggee.ErrorLevel {
		log.Println("error", msg)
	}
}

// Fatal writes a fatal error message and exits.
func (std *stdLog) Fatal(msg string) {
	log.Fatalln("error", msg)
}

// Debugf writes a formated debug message.
func (std *stdLog) Debugf(fmtStr string, args ...interface{}) {
	if std.l == loggee.DebugLevel {
		log.Println("debug", fmt.Sprintf(fmtStr, args...))
	}
}

// Infof writes a formated info message.
func (std *stdLog) Infof(fmtStr string, args ...interface{}) {
	if std.l <= loggee.InfoLevel {
		log.Println("info", fmt.Sprintf(fmtStr, args...))
	}
}

// Warnf writes a formated warn message.
func (std *stdLog) Warnf(fmtStr string, args ...interface{}) {
	if std.l <= loggee.WarnLevel {
		log.Println("warn", fmt.Sprintf(fmtStr, args...))
	}
}

// Errorf writes a formated error message.
func (std *stdLog) Errorf(fmtStr string, args ...interface{}) {
	if std.l <= loggee.ErrorLevel {
		log.Println("error", fmt.Sprintf(fmtStr, args...))
	}
}

// Fatalf writes a formated fatal message and exits.
func (std *stdLog) Fatalf(fmtStr string, args ...interface{}) {
	log.Fatalln("error", fmt.Sprintf(fmtStr, args...))
}
