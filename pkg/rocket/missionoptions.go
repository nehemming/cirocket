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

package rocket

import (
	"fmt"

	"github.com/nehemming/cirocket/pkg/loggee"
)

type missionOptionLog struct {
	log loggee.Logger
}

func (missionOptionLog) Name() string { return "log" }

// LoggerOption sets specifies the logger to use with the mission control.
func LoggerOption(log loggee.Logger) Option {
	return missionOptionLog{log}
}

func (mc *missionControl) SetOptions(options ...Option) error {
	for _, opt := range options {
		switch option := opt.(type) {
		case missionOptionLog:
			mc.log = option.log
		default:
			return fmt.Errorf("option %s not supported", opt.Name())
		}
	}

	return nil
}
