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
	"context"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestListTypes(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	l, err := mc.ListTaskTypes(context.Background())
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(l) != 1 {
		t.Error("len", len(l))
	}
}

func TestListBlueprintsNoSources(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	l, err := mc.ListBlueprints(context.Background(), []string{})
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(l) != 0 {
		t.Error("len", len(l))
	}
}

func TestListBlueprintsTestData(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	mc := NewMissionControl()

	l, err := mc.ListBlueprints(context.Background(), []string{"testdata", filepath.FromSlash("testdata/more")})
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(l) != 2 {
		t.Error("len", len(l))
		return
	}

	if l[0].Name != "blue" {
		t.Error("expected blue, got", l[0].Name)
	}
}
