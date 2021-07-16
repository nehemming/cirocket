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
