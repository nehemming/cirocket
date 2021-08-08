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

package loggee

import (
	"context"
	"errors"
	"testing"
	"time"
)

type (
	testLog struct {
		t     *testing.T
		level Level
	}
)

func newTesLogger(t *testing.T) *testLog { //nolint
	return &testLog{t: t, level: DebugLevel}
}

func TestSetLoggerPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		}
	}()
	SetLogger(nil)
}

func TestActivity(t *testing.T) {
	tl := newTesLogger(t)

	SetLogger(tl)

	ctx := context.Background()

	err := Activity(ctx, func(tc context.Context) error {
		if tc != ctx {
			t.Error("context mismatch")
		}

		return errors.New("test123")
	})

	if err == nil || err.Error() != "test123" {
		t.Error("Error mismatch", err)
	}
}

// Activity is a nested activity.
func (tLog *testLog) Activity(ctx context.Context, fn ActivityFunc) error {
	return fn(ctx)
}

func TestSetLevel(t *testing.T) {
	tl := newTesLogger(t)

	if tl.level != 0 {
		t.Error("default level wrong", tl.level)
	}
	SetLogger(tl)

	tl.SetLevel(ErrorLevel)

	if tl.level != ErrorLevel {
		t.Error("set level wrong", tl.level)
	}
}

func (tLog *testLog) SetLevel(l Level) {
	tLog.level = l
}

func TestWithFields(t *testing.T) {
	tl := newTesLogger(t)

	SetLogger(tl)

	ent := WithFields(Fields{"test": "10"})

	tEnt := ent.(*entry)

	if tEnt.fields.Get("test").(string) != "10" {
		t.Error("Bad With Fields")
	}
}

// WithFields adds fields to a log entry.
func (tLog *testLog) WithFields(f Fielder) Entry {
	return newEntry().WithFields(f)
}

func TestWithField(t *testing.T) {
	tl := newTesLogger(t)

	SetLogger(tl)

	ent := WithField("test", "10")

	tEnt := ent.(*entry)

	if tEnt.fields.Get("test").(string) != "10" {
		t.Error("Bad With Fields")
	}
}

// WithField adds a field to an Entry.
func (tLog *testLog) WithField(key string, value interface{}) Entry {
	return newEntry().WithField(key, value)
}

func TestWithDuration(t *testing.T) {
	tl := newTesLogger(t)

	SetLogger(tl)

	ent := WithDuration(10 * time.Second)

	tEnt := ent.(*entry)

	if tEnt.durattion == nil || *tEnt.durattion != (10*time.Second) {
		t.Error("Bad durations", tEnt.durattion)
	}
}

// WithDuration adds a time t an entry.
func (tLog *testLog) WithDuration(d time.Duration) Entry {
	return newEntry().WithDuration(d)
}

func TestWithError(t *testing.T) {
	tl := newTesLogger(t)

	SetLogger(tl)

	ent := WithError(errors.New("with error"))

	tEnt := ent.(*entry)

	if tEnt.err == nil || tEnt.err.Error() != "with error" {
		t.Error("Bad WithError", tEnt.err)
	}
}

// WithError adds an error to the entry.
func (tLog *testLog) WithError(e error) Entry {
	return newEntry().WithError(e)
}

func TestDebug(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Debug("debug")
}

// Debug writes a debug message.
func (tLog *testLog) Debug(msg string) {
	if msg != "debug" {
		tLog.t.Error("Bad debug", msg)
	}
}

func TestInfo(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Info("info")
}

// Info writes a info message.
func (tLog *testLog) Info(msg string) {
	if msg != "info" {
		tLog.t.Error("Bad info", msg)
	}
}

func TestWarn(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Warn("warn")
}

// Warn writes a warning message.
func (tLog *testLog) Warn(msg string) {
	if msg != "warn" {
		tLog.t.Error("Bad warn", msg)
	}
}

func TestError(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Error("error")
}

// Error writes an error message.
func (tLog *testLog) Error(msg string) {
	if msg != "error" {
		tLog.t.Error("Bad error", msg)
	}
}

func TestFatal(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Fatal("fatal")
}

// Fatal writes a fatal error message and exits.
func (tLog *testLog) Fatal(msg string) {
	if msg != "fatal" {
		tLog.t.Error("Bad fatal", msg)
	}
}

func TestDebugF(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Debugf("debug %s", "debug")
}

// Debugf writes a formated debug message.
func (tLog *testLog) Debugf(fmt string, args ...interface{}) {
	if fmt != "debug %s" {
		tLog.t.Error("Bad debugf", fmt)
	}
}

func TestInfoF(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Infof("info %s", "info")
}

// Infof writes a formated info message.
func (tLog *testLog) Infof(fmt string, args ...interface{}) {
	if fmt != "info %s" {
		tLog.t.Error("Bad infof", fmt)
	}
}

func TestWarnF(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Warnf("warn %s", "warn")
}

// Warnf writes a formated warn message.
func (tLog *testLog) Warnf(fmt string, args ...interface{}) {
	if fmt != "warn %s" {
		tLog.t.Error("Bad warnf", fmt)
	}
}

func TestErrorF(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Errorf("error %s", "error")
}

// Errorf writes a formated error message.
func (tLog *testLog) Errorf(fmt string, args ...interface{}) {
	if fmt != "error %s" {
		tLog.t.Error("Bad errorf", fmt)
	}
}

func TestFatalF(t *testing.T) {
	tl := newTesLogger(t)
	SetLogger(tl)
	Fatalf("fatal %s", "fatal")
}

// Fatalf writes a formated fatal message and exits.
func (tLog *testLog) Fatalf(fmt string, args ...interface{}) {
	if fmt != "fatal %s" {
		tLog.t.Error("Bad fatal", fmt)
	}
}
