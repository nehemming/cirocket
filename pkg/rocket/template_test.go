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
	"runtime"
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestIndentTemplateFunc(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	s, err := capComm.ExpandString(ctx, "spacing", `{{"\nhello"|indent 6}}
	`)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if !strings.HasPrefix(s, "\n      hello") {
		t.Error("indent missing", s)
	}
}

func TestIndentFirstLineTemplateFunc(t *testing.T) {
	ctx := context.Background()
	capComm := NewCapComm(testMissionFile, stdlog.New())

	s, err := capComm.ExpandString(ctx, "spacing", `{{"hello"|indent 6}}
	`)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if !strings.HasPrefix(s, "hello") {
		t.Error("indent present", s)
	}
}

func TestUltimate(t *testing.T) {
	if runtime.GOOS == "windows" {
		return // skip as covered in resources
	}

	d, e := ultimate("/a/b/c")

	if d != "file:///a/b/c" || e != nil {
		t.Error("unexpected", d, e)
	}
}

func TestBUltimateURL(t *testing.T) {
	if d, e := ultimate("file:///a/b/c"); d != "file:///a/b/c" || e != nil {
		t.Error("unexpected", d, e)
	}
}

func TestBUltimateBadURL(t *testing.T) {
	if d, e := ultimate("https://a:a/b:/c"); d != "" || e == nil {
		t.Error("unexpected", d, e)
	}
}

func TestUltimateFilePath(t *testing.T) {
	if d, e := ultimate(filepath.Join("a", "b", "c.txt")); !strings.HasSuffix(d, "/a/b/c.txt") || e != nil {
		t.Error("unexpected", d, e)
	}
}

func TestDirName(t *testing.T) {
	if d := dirname("a/b/c"); d != "a/b" {
		t.Error("unexpected", d)
	}
}

func TestDirNameURL(t *testing.T) {
	if d := dirname("file:///a/b/c"); d != "file:/a/b" {
		t.Error("unexpected", d)
	}
}

func TestDirNameFilePath(t *testing.T) {
	if d := dirname(filepath.Join("a", "b", "c")); d != "a/b" {
		t.Error("unexpected", d)
	}
}

func TestBaseName(t *testing.T) {
	if d := basedname("a/b/c"); d != "c" {
		t.Error("unexpected", d)
	}
}

func TestBaseNameURL(t *testing.T) {
	if d := basedname("file:///a/b/c"); d != "c" {
		t.Error("unexpected", d)
	}
}

func TestBaseNameFilePath(t *testing.T) {
	if d := basedname(filepath.Join("a", "b", "c.txt")); d != "c.txt" {
		t.Error("unexpected", d)
	}
}
