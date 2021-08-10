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

package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
)

func TestMoveRun(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	dir := filepath.Join("testdata", "mv")
	_ = os.RemoveAll(filepath.FromSlash(dir))
	defer func() { _ = os.RemoveAll(filepath.FromSlash(dir)) }()

	mission, cfgFile := loadMission("move")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	validateMoveTest(t, dir)
}

func validateMoveTest(t *testing.T, dir string) {
	t.Helper()
	// Check and clean
	expecttedFromSource, _ := globFileAbsRel("*/*.yml", "*.go")

	stage, _ := globFileAbsRel(filepath.Join(dir, "s/**"))
	dest, _ := globFileAbsRel(filepath.Join(dir, "d/**"))
	keep, _ := globFileAbsRel(filepath.Join(dir, "k/**"))

	if len(expecttedFromSource) != len(dest) {
		t.Errorf("File mismatch dest wanted %d got %d", len(expecttedFromSource), len(dest))
	}
	if len(expecttedFromSource) != len(keep) {
		t.Errorf("File mismatch keep wanted %d got %d", len(expecttedFromSource), len(keep))
	}
	if 0 != len(stage) {
		t.Errorf("File mismatch stage wanted 0 got %d", len(stage))
	}
}
