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

// Package apexlog implements an apex logger wrapper for loggee.Logger
package apexlog

import (
	"context"
	"time"

	"github.com/apex/log"
	logcli "github.com/apex/log/handlers/cli"
	"github.com/nehemming/cirocket/pkg/loggee"
)

const ctxKey = ctxKeyType("apex")

type (
	ctxKeyType string

	apexLogger int
)

// New creates a new loggee logger from an apex log handler.
func New(h log.Handler) loggee.Logger {
	log.SetHandler(h)

	return apexLogger(0)
}

// NewCli creates a cli logger.
func NewCli(padding int) loggee.Logger {
	logcli.Default.Padding = padding

	log.SetHandler(logcli.Default)

	return apexLogger(padding)
}

// Activity is a nested activity.
func (a apexLogger) Activity(ctx context.Context, fn loggee.ActivityFunc) error {
	depth, _ := ctx.Value(ctxKey).(int)
	padding := logcli.Default.Padding
	logcli.Default.Padding = (1 + depth) * int(a)

	err := fn(context.WithValue(ctx, ctxKey, depth+1))
	if err != nil {
		// Exit without resetting context, logging will be at this level
		return err
	}

	logcli.Default.Padding = padding
	return nil
}

// SetLevel sets the logging level.
func (a apexLogger) SetLevel(l loggee.Level) {
	log.SetLevel(log.Level(l))
}

// WithFields adds fields to a log entry.
func (a apexLogger) WithFields(f loggee.Fielder) loggee.Entry {
	fields := make(log.Fields)

	for k, v := range f.Fields() {
		fields[k] = v
	}

	return wrapper{log.WithFields(fields)}
}

// WithField adds a field to an Entry.
func (a apexLogger) WithField(key string, value interface{}) loggee.Entry {
	return wrapper{log.WithField(key, value)}
}

// WithDuration adds a time t an Entry.
func (a apexLogger) WithDuration(d time.Duration) loggee.Entry {
	return wrapper{log.WithDuration(d)}
}

// WithError adds an error to the Entry.
func (a apexLogger) WithError(e error) loggee.Entry {
	return wrapper{log.WithError(e)}
}

// Debug writes a debug message.
func (a apexLogger) Debug(msg string) {
	log.Debug(msg)
}

// Info writes a info message.
func (a apexLogger) Info(msg string) {
	log.Info(msg)
}

// Warn writes a warning message.
func (a apexLogger) Warn(msg string) {
	log.Warn(msg)
}

// Error writes an error message.
func (a apexLogger) Error(msg string) {
	log.Error(msg)
}

// Fatal writes a fatal error message and exits.
func (a apexLogger) Fatal(msg string) {
	log.Fatal(msg)
}

// Debugf writes a formated debug message.
func (a apexLogger) Debugf(fmt string, args ...interface{}) {
	log.Debugf(fmt, args...)
}

// Infof writes a formated info message.
func (a apexLogger) Infof(fmt string, args ...interface{}) {
	log.Infof(fmt, args...)
}

// Warnf writes a formated warn message.
func (a apexLogger) Warnf(fmt string, args ...interface{}) {
	log.Warnf(fmt, args...)
}

// Errorf writes a formated error message.
func (a apexLogger) Errorf(fmt string, args ...interface{}) {
	log.Errorf(fmt, args...)
}

// Fatalf writes a formated fatal message and exits.
func (a apexLogger) Fatalf(fmt string, args ...interface{}) {
	log.Fatalf(fmt, args...)
}
