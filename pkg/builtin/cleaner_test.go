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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
	"github.com/nehemming/cirocket/pkg/rocket"
)

func TestCleanerType(t *testing.T) {
	var rt cleanerType

	if rt.Type() != "cleaner" {
		t.Error("Wrong cleaner type", rt.Type())
	}
}

func TestCleanerDesc(t *testing.T) {
	var rt cleanerType

	if rt.Description() == "" {
		t.Error("needs description", rt.Type())
	}
}

func testCounts(t *testing.T, i int, files []string, dir string) {
	t.Helper()
	count := len(files)

	if i == 0 {
		if count != 5 {
			t.Error("dir mismatch", dir, count)
		}
	} else if i == 1 {
		if count != 4 {
			t.Error("dir mismatch", dir, count)
		}
	} else if count != 0 {
		t.Error("dir mismatch", dir, count)
	}
}

func validateCleanerTest(t *testing.T) {
	t.Helper()
	// Check and clean
	for i := 0; i < 5; i++ {
		dir := fmt.Sprintf("%s-%d", filepath.Join("testdata", "clean"), i)

		runbook := filepath.Join(dir, "*")

		if i != 3 {
			// need 5 files
			if _, err := os.Stat(filepath.FromSlash(dir)); err != nil {
				t.Error("dir missing", dir, err)
			}

			files, err := filepath.Glob(runbook)
			if err != nil {
				t.Error("error listing", dir, err)
			}

			testCounts(t, i, files, dir)
		} else if _, err := os.Stat(filepath.FromSlash(dir)); err == nil {
			t.Error("dir present", dir)
		}
		_ = os.RemoveAll(filepath.FromSlash(dir))
	}
}

func TestCleanerRun(t *testing.T) {
	removeRun(t, "cleaner")
}

func TestRemoveRun(t *testing.T) {
	removeRun(t, "remove")
}

func removeRun(t *testing.T, missionName string) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	// create some data
	for i := 0; i < 5; i++ {
		dir := fmt.Sprintf("%s-%d", filepath.Join("testdata", "clean"), i)
		// check clean
		_ = os.RemoveAll(filepath.FromSlash(dir))

		if err := os.MkdirAll(filepath.FromSlash(dir), 0o777); err != nil {
			panic(err)
		}

		for f := 0; f < 5; f++ {
			name := fmt.Sprintf("file-%d", f)
			fn := filepath.Join(dir, name)
			if err := os.WriteFile(fn, []byte("hello"), 0o666); err != nil {
				panic(err)
			}
		}
	}

	mission, cfgFile := loadMission(missionName)

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	validateCleanerTest(t)
}

func TestDeleteMissing(t *testing.T) {
	files := []string{"filenotexists.gogo"}

	err := deleteFiles(files, nil)
	if err != nil {
		t.Error("unexpected", err)
	}
}
