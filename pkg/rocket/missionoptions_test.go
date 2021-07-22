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
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestMissionOptionLogName(t *testing.T) {
	ml := (missionOptionLog{})
	if ml.Name() != "log" {
		t.Error("unexpectedname", ml.Name())
	}
}

func TestLoggerOption(t *testing.T) {
	l := stdlog.New()
	opt := LoggerOption(l)

	if option, ok := opt.(missionOptionLog); !ok {
		t.Error("not missionOptionLog", opt)
	} else if option.log != l {
		t.Error("wrong log")
	}
}

func TestSetOptionsLog(t *testing.T) {
	mc := NewMissionControl()
	l := stdlog.New()

	err := mc.SetOptions(LoggerOption(l))
	if err != nil {
		t.Error("unexpected", err)
	}

	if mc.(*missionControl).log != l {
		t.Error("unexpected logger")
	}
}

type unknownOption struct{}

func (unknownOption) Name() string { return "unknown" }

func TestSetOptionsUnknownErrors(t *testing.T) {
	mc := NewMissionControl()

	err := mc.SetOptions(unknownOption{})
	if err == nil {
		t.Error("expected error")
	}
}
