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

package apexlog

import (
	"time"

	"github.com/apex/log"
	"github.com/nehemming/cirocket/pkg/loggee"
)

type wrapper struct {
	entry *log.Entry
}

// WithFields adds fields to a log Entry.
func (w wrapper) WithFields(f loggee.Fielder) loggee.Entry {
	fields := make(log.Fields)

	for k, v := range f.Fields() {
		fields[k] = v
	}

	return wrapper{w.entry.WithFields(fields)}
}

// WithField adds a field to an Entry.
func (w wrapper) WithField(key string, value interface{}) loggee.Entry {
	return wrapper{w.entry.WithField(key, value)}
}

// WithDuration adds a time t an Entry.
func (w wrapper) WithDuration(d time.Duration) loggee.Entry {
	return wrapper{w.entry.WithDuration(d)}
}

// WithError adds an error to the Entry.
func (w wrapper) WithError(e error) loggee.Entry {
	return wrapper{w.entry.WithError(e)}
}

// Debug writes a debug message.
func (w wrapper) Debug(msg string) {
	w.entry.Debug(msg)
}

// Info writes a info message.
func (w wrapper) Info(msg string) {
	w.entry.Info(msg)
}

// Warn writes a warning message.
func (w wrapper) Warn(msg string) {
	w.entry.Warn(msg)
}

// Error writes an error message.
func (w wrapper) Error(msg string) {
	w.entry.Error(msg)
}

// Fatal writes a fatal error message and exits.
func (w wrapper) Fatal(msg string) {
	w.entry.Fatal(msg)
}

// Debugf writes a formated debug message.
func (w wrapper) Debugf(fmt string, args ...interface{}) {
	w.entry.Debugf(fmt, args...)
}

// Infof writes a formated info message.
func (w wrapper) Infof(fmt string, args ...interface{}) {
	w.entry.Infof(fmt, args...)
}

// Warnf writes a formated warn message.
func (w wrapper) Warnf(fmt string, args ...interface{}) {
	w.entry.Warnf(fmt, args...)
}

// Errorf writes a formated error message.
func (w wrapper) Errorf(fmt string, args ...interface{}) {
	w.entry.Errorf(fmt, args...)
}

// Fatalf writes a formated fatal message and exits.
func (w wrapper) Fatalf(fmt string, args ...interface{}) {
	w.entry.Fatalf(fmt, args...)
}
