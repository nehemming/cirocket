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

import (
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewLogProviderSeverityError(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogError)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityWarn(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogWarn)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityInfo(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, LogInfo)

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}

func TestNewLogProviderSeverityUndefined(t *testing.T) {
	log := stdlog.New()

	lp := NewLogProvider(log, Severity(99))

	writer, err := lp.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
	}
}
