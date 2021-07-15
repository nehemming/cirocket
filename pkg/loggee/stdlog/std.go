package stdlog

import (
	"context"
	"time"

	"github.com/apex/log"
	"github.com/nehemming/cirocket/pkg/loggee"
)

type (
	stdLog int
)

func New() loggee.Logger {

	return stdLog(0)
}

// Activity is a nested activity
func (std stdLog) Activity(ctx context.Context, fn loggee.ActivityFunc) error {
	return fn(ctx)
}

// WithFields adds fields to a log entry
func (std stdLog) WithFields(f loggee.Fielder) loggee.Entry {

	return newEntry().WithFields(f)
}

// WithField adds a field to an Entry
func (std stdLog) WithField(key string, value interface{}) loggee.Entry {

	return newEntry().WithField(key, value)
}

// WithDuration adds a time t an entry
func (std stdLog) WithDuration(d time.Duration) loggee.Entry {

	return newEntry().WithDuration(d)
}

// WithError adds an error to the entry
func (std stdLog) WithError(e error) loggee.Entry {

	return newEntry().WithError(e)
}

// Debug writes a debug message
func (std stdLog) Debug(msg string) {

	log.Debug(msg)
}

// Info writes a info message
func (std stdLog) Info(msg string) {

	log.Info(msg)
}

// Warn writes a warning message
func (std stdLog) Warn(msg string) {

	log.Warn(msg)
}

// Error writes an error message
func (std stdLog) Error(msg string) {

	log.Error(msg)
}

// Fatal writes a fatal error message and exits
func (std stdLog) Fatal(msg string) {

	log.Fatal(msg)
}

// Debugf writes a formated debug message
func (std stdLog) Debugf(fmt string, args ...interface{}) {

	log.Debugf(fmt, args...)
}

// Infof writes a formated info message
func (std stdLog) Infof(fmt string, args ...interface{}) {

	log.Infof(fmt, args...)
}

// Warnf writes a formated warn message
func (std stdLog) Warnf(fmt string, args ...interface{}) {

	log.Warnf(fmt, args...)
}

// Errorf writes a formated error message
func (std stdLog) Errorf(fmt string, args ...interface{}) {

	log.Errorf(fmt, args...)
}

// Fatalf writes a formated fatal message and exits
func (std stdLog) Fatalf(fmt string, args ...interface{}) {

	log.Fatalf(fmt, args...)
}
