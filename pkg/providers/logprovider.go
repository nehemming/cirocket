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

package providers

import "github.com/nehemming/cirocket/pkg/loggee"

const (
	// LogInfo is an informational level of logging.
	LogInfo = Severity(iota)
	// LogWarn is a warning level of logging.
	LogWarn
	// LogError is an error level of logging.
	LogError
)

// Severity is the level to log messages at in a log provider.
type Severity int

// NewLogProvider creates provider that writes to a log.
func NewLogProvider(log loggee.Logger, severity Severity) ResourceProvider {
	var logFunc LineFunc

	switch severity {
	case LogInfo:
		logFunc = log.Info
	case LogWarn:
		logFunc = log.Warn
	case LogError:
		logFunc = log.Error
	default:
		logFunc = func(_ string) {}
	}

	return logFunc
}
