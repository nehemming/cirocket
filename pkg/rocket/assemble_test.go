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
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestAssembleEmptySources(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{}

	err := mc.Assemble(ctx, "none", sources, "", nil)
	if err == nil || err.Error() != "blueprint none: cannot find in sources" {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSources(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "none", sources, "", nil)
	if err == nil || err.Error() != "blueprint none: cannot find in sources" {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSourcesBlueNoMission(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "blue", sources, "", nil)
	if err == nil || filepath.ToSlash(err.Error()) != "loading mission for blueprint blue (testdata/blue): no source was specified" {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSourcesBlueNoType(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()

	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "blue_inline", sources, "", nil)
	if err == nil || err.Error() != "testing prepare: prepare: task to test: unknown task type testTask" {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSourcesInlineMission(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "blue_inline", sources, "", nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestAssembleAbsoluteSourceInlineMission(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	absBlueprint := filepath.Join(cwd, "testdata", "blue_inline")

	err = mc.Assemble(ctx, absBlueprint, sources, "", nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestAssembleAbsoluteFileInlineMission(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	absBlueprint := filepath.Join(cwd, "testdata", "blue_inline", "blueprint.yml")

	err = mc.Assemble(ctx, absBlueprint, sources, "", nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSourcesPathMission(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "blue_runbook", sources, "", nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestAssembleOneSourcesPathMissionWithSpec(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{"testdata"}

	spedLocation := filepath.Join("testdata", "more", "blue_runbook_in.yml")

	err := mc.Assemble(ctx, "blue_runbook", sources, spedLocation, nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestAssembleRunbookErrors(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	tt := &testTaskType{t: t}
	mc.RegisterTaskTypes(tt)

	sources := []string{"testdata"}

	err := mc.Assemble(ctx, "blue_runbook", sources, "bad", nil)
	if err == nil {
		t.Error("error unexpected")
	}
}

func TestLoadMapFromLocationDecodeFail(t *testing.T) {
	ctx := context.Background()
	m, n, err := loadMapFromLocation(ctx, Location{Inline: "bad"}, "")
	if err == nil || n != "" || m != nil {
		t.Error("error unexpected", err, n, m)
	}
}
