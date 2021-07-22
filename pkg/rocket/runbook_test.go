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
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestGetRunbookNoSource(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{}

	_, err := mc.GetRunbook(ctx, "blue", sources)
	if err == nil || err.Error() != "blueprint blue: cannot find in sources" {
		t.Error("unexpected", err)
	}
}

func TestGetRunbookCannotFindBlueprint(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	_, err := mc.GetRunbook(ctx, "noone", sources)
	if err == nil || err.Error() != "blueprint noone: cannot find in sources" {
		t.Error("unexpected", err)
	}
}

func TestGetRunbookBlankSpec(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	runbook, err := mc.GetRunbook(ctx, "blue", sources)
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if runbook != "" {
		t.Error("runbook", runbook)
	}
}

func TestGetRunbookInlineSpec(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	runbook, err := mc.GetRunbook(ctx, "blue_inline", sources)
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if !strings.HasPrefix(runbook, "# comment preserved") {
		t.Error("runbook", runbook)
	}
}

func TestGetRunbookInlineSpecNotFirstHit(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"missing", "testdata"}

	runbook, err := mc.GetRunbook(ctx, "blue_inline", sources)
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if !strings.HasPrefix(runbook, "# comment preserved") {
		t.Error("runbook", runbook)
	}
}

func TestGetRunbookSideFile(t *testing.T) {
	loggee.SetLogger(stdlog.New())
	ctx := context.Background()
	mc := NewMissionControl()
	sources := []string{"testdata"}

	runbook, err := mc.GetRunbook(ctx, "blue_runbook", sources)
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if !strings.HasPrefix(runbook, "# comment preserved") {
		t.Error("runbook", runbook)
	}
}
