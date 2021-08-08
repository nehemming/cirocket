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

// Package loggee is logging wrapper interface that allows share packages to call logging services without
// the calling application being tide to a specific implementation.
package loggee

import (
	"context"
	"time"
)

// Log levels.
const (
	InvalidLevel Level = iota - 1
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type (
	// ActivityFunc is the call back function signature for an activity
	// Activities allow logging context to be added to output.
	ActivityFunc func(context.Context) error

	// Entry is a wrapper interface for a slog entry.
	Entry interface {
		WithFields(Fielder) Entry
		WithField(string, interface{}) Entry
		WithDuration(time.Duration) Entry
		WithError(error) Entry
		Debug(string)
		Info(string)
		Warn(string)
		Error(string)
		Fatal(string)
		Debugf(string, ...interface{})
		Infof(string, ...interface{})
		Warnf(string, ...interface{})
		Errorf(string, ...interface{})
		Fatalf(string, ...interface{})
	}

	// Level of severity.
	Level int

	// Logger is the generic interface for all loggers.
	Logger interface {

		// Entry implements the main logging interface.
		Entry

		// Activity log within th context of the activity.
		Activity(ctx context.Context, fn ActivityFunc) error

		SetLevel(l Level)
	}
)

var levelNames = [...]string{
	DebugLevel: "debug",
	InfoLevel:  "info",
	WarnLevel:  "warn",
	ErrorLevel: "error",
	FatalLevel: "fatal",
}

// String implementation.
func (l Level) String() string {
	return levelNames[l]
}

var defaultLog Logger

// SetLogger sets the logger.
func SetLogger(logger Logger) {
	defaultLog = logger
	mustHaveLogger()
}

// Default returns the default logger, if none is set Default will panic.
func Default() Logger {
	mustHaveLogger()
	return defaultLog
}

// mustHaveLogger checks that a logger has been set. If there is no logger this method will panic.
func mustHaveLogger() {
	if defaultLog == nil {
		panic("no logger has been provided, call loggee.SetLogger with a non nil logger")
	}
}

// Activity log within th context of the activity.
func Activity(ctx context.Context, fn ActivityFunc) error {
	mustHaveLogger()

	return defaultLog.Activity(ctx, fn)
}

// WithFields adds fields to a log Entry.
func WithFields(f Fielder) Entry {
	mustHaveLogger()

	return defaultLog.WithFields(f)
}

// WithField adds a field to an Entry.
func WithField(key string, value interface{}) Entry {
	mustHaveLogger()

	return defaultLog.WithField(key, value)
}

// WithDuration adds a time t an Entry.
func WithDuration(d time.Duration) Entry {
	mustHaveLogger()

	return defaultLog.WithDuration(d)
}

// WithError adds an error to the Entry.
func WithError(e error) Entry {
	mustHaveLogger()

	return defaultLog.WithError(e)
}

// Debug writes a debug message.
func Debug(msg string) {
	mustHaveLogger()

	defaultLog.Debug(msg)
}

// Info writes a info message.
func Info(msg string) {
	mustHaveLogger()

	defaultLog.Info(msg)
}

// Warn writes a warning message.
func Warn(msg string) {
	mustHaveLogger()

	defaultLog.Warn(msg)
}

// Error writes an error message.
func Error(msg string) {
	mustHaveLogger()

	defaultLog.Error(msg)
}

// Fatal writes a fatal error message and exits.
func Fatal(msg string) {
	mustHaveLogger()

	defaultLog.Fatal(msg)
}

// Debugf writes a formated debug message.
func Debugf(fmt string, args ...interface{}) {
	mustHaveLogger()

	defaultLog.Debugf(fmt, args...)
}

// Infof writes a formated info message.
func Infof(fmt string, args ...interface{}) {
	mustHaveLogger()

	defaultLog.Infof(fmt, args...)
}

// Warnf writes a formated warn message.
func Warnf(fmt string, args ...interface{}) {
	mustHaveLogger()

	defaultLog.Warnf(fmt, args...)
}

// Errorf writes a formated error message.
func Errorf(fmt string, args ...interface{}) {
	mustHaveLogger()

	defaultLog.Errorf(fmt, args...)
}

// Fatalf writes a formated fatal message and exits.
func Fatalf(fmt string, args ...interface{}) {
	mustHaveLogger()

	defaultLog.Fatalf(fmt, args...)
}
