package loggee

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type entry struct {
	fields    Fields
	durattion *time.Duration
	err       error
}

func (ent *entry) String() string {
	var b strings.Builder

	for k, v := range ent.fields {
		b.WriteString(fmt.Sprintf("%s:%v\t", k, v))
	}

	if ent.durattion != nil {
		b.WriteString(fmt.Sprintf("duration:%v\t", *ent.durattion))
	}

	if ent.err != nil {
		b.WriteString(fmt.Sprintf("error:%v\t", ent.err))
	}

	return b.String()
}

func newEntry() *entry {
	return &entry{
		fields: make(Fields),
	}
}

// WithFields adds fields to a log Entry.
func (ent *entry) WithFields(f Fielder) Entry {
	for k, v := range f.Fields() {
		ent.fields[k] = v
	}

	return ent
}

// WithField adds a field to an Entry.
func (ent *entry) WithField(key string, value interface{}) Entry {
	ent.fields[key] = value

	return ent
}

// WithDuration adds a time t an Entry.
func (ent *entry) WithDuration(d time.Duration) Entry {
	ent.durattion = &d

	return ent
}

// WithError adds an error to the Entry.
func (ent *entry) WithError(e error) Entry {
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
