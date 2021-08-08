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

package buildinfo

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/nehemming/testsupport"
)

func TestNewInfo(t *testing.T) {
	bi := NewInfo("v", "c", "d", "b", "cn")
	bi.Architecture = "amd64"
	bi.OperatingSystem = "darwin"

	if bi.BuiltBy != "b" {
		t.Error("Unexpected BuiltBy", bi.BuiltBy)
	}

	if bi.Commit != "c" {
		t.Error("Unexpected Commit", bi.Commit)
	}

	if bi.CompiledName != "cn" {
		t.Error("Unexpected CompiledName", bi.CompiledName)
	}

	if bi.Date != "d" {
		t.Error("Unexpected Date", bi.Date)
	}

	if bi.Version != "v" {
		t.Error("Unexpected Version", bi.Version)
	}

	if bi.RunName == "" {
		t.Error("Unexpected RunName", bi.RunName)
	}

	if !strings.HasSuffix(bi.StartDir, "buildinfo") {
		t.Error("Unexpected StartDir", bi.StartDir)
	}

	if bi.String() != "v darwin/amd64 c d cn [b]" {
		t.Error("Unexpected String", bi.String())
	}
}

func TestContextRoundTrip(t *testing.T) {
	bi := NewInfo("v", "c", "d", "b", "cn")

	ctx := bi.NewContext(context.Background())

	ret := GetBuildInfo(ctx)

	if ret.BuiltBy != bi.BuiltBy {
		t.Error("Unexpected BuiltBy", ret.BuiltBy)
	}
	if ret.Commit != bi.Commit {
		t.Error("Unexpected Commit", ret.Commit)
	}
	if ret.CompiledName != bi.CompiledName {
		t.Error("Unexpected CompiledName", ret.CompiledName)
	}
	if ret.Date != bi.Date {
		t.Error("Unexpected Date", ret.Date)
	}
	if ret.RunName != bi.RunName {
		t.Error("Unexpected RunName", ret.RunName)
	}
	if ret.StartDir != bi.StartDir {
		t.Error("Unexpected StartDir", ret.StartDir)
	}
	if ret.Version != bi.Version {
		t.Error("Unexpected Version", ret.Version)
	}
}

func TestContextRuntime(t *testing.T) {
	bi := NewInfo("v", "c", "d", "b", "cn")

	ctx := bi.NewContext(context.Background())

	ret := GetBuildInfo(ctx)

	if ret.Architecture != runtime.GOARCH {
		t.Error("Unexpected Architecture", ret.Architecture)
	}
	if ret.OperatingSystem != runtime.GOOS {
		t.Error("Unexpected OperatingSystem", ret.Version)
	}
	if ret.Compiler != runtime.Version() {
		t.Error("Unexpected Compiler", ret.Compiler)
	}
}

func TestStringVersion(t *testing.T) {
	bi := NewInfo("v", "c", "d", "b", "cn")
	bi.Architecture = "amd64"
	bi.OperatingSystem = "darwin"
	if r := bi.String(); r != "v darwin/amd64 c d cn [b]" {
		t.Error("unexpected", r)
	}
}

func TestBasicVersion(t *testing.T) {
	bi := NewInfo("v", "c", "d", "b", "cn")
	bi.Architecture = "amd64"
	bi.OperatingSystem = "darwin"

	if r := bi.BasicVersion(); r != "v darwin/amd64 c d cn [b]" {
		t.Error("unexpected", r)
	}
}

func TestBasicVersionLongCommitHash(t *testing.T) {
	bi := NewInfo("v", "c23456789", "d", "b", "cn")
	bi.Architecture = "amd64"
	bi.OperatingSystem = "darwin"
	bi.Compiler = "go1.16.6"

	if r := bi.BasicVersion(); r != "v darwin/amd64 c234567 d cn [b]" {
		t.Error("unexpected", r)
	}
}

func TestTabularVersion(t *testing.T) {
	const expected = `        Version: v
   CompiledName: cn
   Architecture: amd64
OperatingSystem: darwin
		   Date: d
	     Commit: c
	    BuiltBy: b
	   Compiler: go1.16.6`

	bi := NewInfo("v", "c", "d", "b", "cn")
	// set fixed runtime values for test
	bi.Architecture = "amd64"
	bi.OperatingSystem = "darwin"
	bi.Compiler = "go1.16.6"

	testsupport.CompareStrings(t, expected, bi.TabularVersion())
}
