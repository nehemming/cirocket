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

func TestCopyType(t *testing.T) {
	var ct copyType

	if ct.Type() != "copy" {
		t.Error("Wrong copy type", ct.Type())
	}
}

func TestGlobFileAbsRel(t *testing.T) {
	files, err := globFileAbsRel("c*.go")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if len(files) != 4 {
		t.Error("unexpected", files, len(files))
		return
	}

	// check has abs path with package as parent dir
	expected := filepath.Base(filepath.Dir(files[0].Abs))
	if expected != "builtin" {
		t.Error("abs", files[0].Abs, expected)
	}

	if files[0].Rel != "cleaner.go" {
		t.Error("rel", files[0].Rel)
	}
}

func TestGlobFileAbsRelSingle(t *testing.T) {
	files, err := globFileAbsRel("cleaner.go")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if len(files) != 1 {
		t.Error("unexpected", files, len(files))
		return
	}

	// check has abs path with package as parent dir
	expected := filepath.Base(filepath.Dir(files[0].Abs))
	if expected != "builtin" {
		t.Error("abs", files[0].Abs, expected)
	}

	if files[0].Rel != "cleaner.go" {
		t.Error("rel", files[0].Rel)
	}
}

func TestGlobFileAbsRelDeep(t *testing.T) {
	files, err := globFileAbsRel("**/*.yml")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	if len(files) != 10 {
		t.Error("unexpected len", len(files), files)
		return
	}

	// check has abs path with package as parent dir
	expected := filepath.Base(filepath.Dir(filepath.Dir(files[0].Abs)))
	if expected != "builtin" {
		t.Error("abs", files[0].Abs, expected)
	}

	r := filepath.ToSlash(files[0].Rel)

	if r != "testdata/badfetch.yml" {
		t.Error("rel", files[0].Rel)
	}
}

func TestGlobFileAbsRelDistinct(t *testing.T) {
	files, err := globFileAbsRel("*.go", "*")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	fc := len(files)

	// don't break for every add src file
	if fc < 14 || fc > 20 {
		t.Error("unexpected len", fc, files)
		return
	}

	// check has abs path with package as parent dir
	expected := filepath.Base(filepath.Dir(files[0].Abs))
	if expected != "builtin" {
		t.Error("abs", files[0].Abs, expected)
	}

	r := filepath.ToSlash(files[0].Rel)

	if r != "cleaner.go" {
		t.Error("rel", files[0].Rel)
	}
}

func TestGetDestSpecBlank(t *testing.T) {
	ds, err := getDestSpec("")

	if err == nil || err.Error() != "destination cannot be blank" {
		t.Error("unexpected", ds, err)
	}
}

func TestGetDestSpec(t *testing.T) {
	ds, err := getDestSpec(".")

	if err != nil || !ds.IsDir {
		t.Error("not a dir", ds, err)
	}

	wd, _ := os.Getwd()

	if ds.Path != wd {
		t.Error("ds", ds)
	}
}

func TestGetDestSpecTestData(t *testing.T) {
	ds, err := getDestSpec("testdata")

	if err != nil || !ds.IsDir {
		t.Error("not a dir", ds, err)
	}

	wd, _ := os.Getwd()

	if ds.Path != filepath.Join(wd, "testdata") {
		t.Error("ds", ds)
	}
}

func TestGetDestSpecFile(t *testing.T) {
	ds, err := getDestSpec("cleaner.go")

	if err != nil || ds.IsDir {
		t.Error("is a dir", ds, err)
	}

	wd, _ := os.Getwd()

	if ds.Path != filepath.Join(wd, "cleaner.go") {
		t.Error("ds", ds)
	}
}

func TestCopyRun(t *testing.T) {
	loggee.SetLogger(stdlog.New())

	mc := rocket.NewMissionControl()
	RegisterAll(mc)

	dir := filepath.Join("testdata", "cpt")
	_ = os.RemoveAll(filepath.FromSlash(dir))
	defer func() { _ = os.RemoveAll(filepath.FromSlash(dir)) }()

	mission, cfgFile := loadMission("copy")

	if err := mc.LaunchMission(context.Background(), cfgFile, mission); err != nil {
		t.Error("failure", err)
	}

	validateCopyTest(t, dir)
}

func validateCopyTest(t *testing.T, dir string) {
	t.Helper()
	// Check and clean
	src, _ := globFileAbsRel("*/*.yml", "*.go")
	dest, _ := globFileAbsRel(filepath.Join(dir, "**"))
	if len(src) != len(dest) {
		t.Errorf("File mismatch wanted %d got %d", len(src), len(dest))
	}
}
