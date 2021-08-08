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

package stdlog

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nehemming/cirocket/pkg/loggee"
)

type entry struct {
	fields    loggee.Fields
	durattion *time.Duration
	err       error
}

func (ent *entry) String() string {
	var b strings.Builder

	for k, v := range ent.fields {
		b.WriteString(fmt.Sprintf("%s:%v ", k, v))
	}

	if ent.durattion != nil {
		b.WriteString(fmt.Sprintf("duration:%v ", *ent.durattion))
	}

	if ent.err != nil {
		b.WriteString(fmt.Sprintf("error:%v ", ent.err))
	}

	return strings.Trim(b.String(), " ")
}

func newEntry() *entry {
	return &entry{
		fields: make(loggee.Fields),
	}
}

// WithFields adds fields to a log Entry.
func (ent *entry) WithFields(f loggee.Fielder) loggee.Entry {
	for k, v := range f.Fields() {
		ent.fields[k] = v
	}

	return ent
}

// WithField adds a field to an Entry.
func (ent *entry) WithField(key string, value interface{}) loggee.Entry {
	ent.fields[key] = value

	return ent
}

// WithDuration adds a time t an Entry.
func (ent *entry) WithDuration(d time.Duration) loggee.Entry {
	ent.durattion = &d

	return ent
}

// WithError adds an error to the Entry.
func (ent *entry) WithError(e error) loggee.Entry {
	ent.err = e

	return ent
}

// Debug writes a debug message.
func (ent *entry) Debug(msg string) {
	log.Println("debug", msg, ent)
}

// Info writes a info message.
func (ent *entry) Info(msg string) {
	log.Println("info", msg, ent)
}

// Warn writes a warning message.
func (ent *entry) Warn(msg string) {
	log.Println("warn", msg, ent)
}

// Error writes an error message.
func (ent *entry) Error(msg string) {
	log.Println("error", msg, ent)
}

// Fatal writes a fatal error message and exits.
func (ent *entry) Fatal(msg string) {
	log.Fatalln("error", msg, ent)
}

// Debugf writes a formated debug message.
func (ent *entry) Debugf(fmtStr string, args ...interface{}) {
	log.Println("debug", fmt.Sprintf(fmtStr, args...), ent)
}

// Infof writes a formated info message.
func (ent *entry) Infof(fmtStr string, args ...interface{}) {
	log.Println("info", fmt.Sprintf(fmtStr, args...), ent)
}

// Warnf writes a formated warn message.
func (ent *entry) Warnf(fmtStr string, args ...interface{}) {
	log.Println("warn", fmt.Sprintf(fmtStr, args...), ent)
}

// Errorf writes a formated error message.
func (ent *entry) Errorf(fmtStr string, args ...interface{}) {
	log.Println("error", fmt.Sprintf(fmtStr, args...), ent)
}

// Fatalf writes a formated fatal message and exits.
func (ent *entry) Fatalf(fmtStr string, args ...interface{}) {
	log.Fatalln(fmt.Sprintf(fmtStr, args...), ent)
}
