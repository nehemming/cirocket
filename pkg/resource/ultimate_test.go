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

package resource

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func TestGetParentLocationEmpty(t *testing.T) {
	u, err := GetParentLocation("")
	if err == nil || u != nil {
		t.Error("expected error")
	}
}

func TestGetParentLocationRoot(t *testing.T) {
	u, err := GetParentLocation("/")
	if err != nil || u == nil || u.String() != "file:///" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationDeep(t *testing.T) {
	u, err := GetParentLocation("/home/work/seen")
	if err != nil || u == nil || u.String() != "file:///home/work" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationDeepRel(t *testing.T) {
	u, err := GetParentLocation("/home/work/seen/../down/..")
	if err != nil || u == nil || u.String() != "file:///home" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationRooted(t *testing.T) {
	u, err := GetParentLocation("/root")
	if err != nil || u == nil || u.String() != "file:///" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationWebRootEmpty(t *testing.T) {
	u, err := GetParentLocation("http://server")
	if err != nil || u == nil || u.String() != "http://server/." {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationWebRootDot(t *testing.T) {
	u, err := GetParentLocation("http://server/.")
	if err != nil || u == nil || u.String() != "http://server/" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationWebRoot(t *testing.T) {
	u, err := GetParentLocation("http://server/root")
	if err != nil || u == nil || u.String() != "http://server/" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationWebDeep(t *testing.T) {
	u, err := GetParentLocation("http://server/home/work/seen")
	if err != nil || u == nil || u.String() != "http://server/home/work" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationDeepWebRel(t *testing.T) {
	u, err := GetParentLocation("http://server/home/work/seen/../down/..")
	if err != nil || u == nil || u.String() != "http://server/home" {
		t.Error("unexpected error", err, u)
	}
}

func TestGetParentLocationBadURL(t *testing.T) {
	u, err := GetParentLocation("http://server::test")
	if err == nil || u != nil {
		t.Error("expected error")
	}
}

func TestGetParentLocationHome(t *testing.T) {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	expected := fmt.Sprintf("file://%s/work", home)

	u, err := GetParentLocation("~/work/seen")
	if err != nil || u == nil || u.String() != expected {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateURLEmpty(t *testing.T) {
	u, err := UltimateURL()
	if err == nil || u != nil {
		t.Error("expected error")
	}
}

func TestUltimateURLMultiFile(t *testing.T) {
	u, err := UltimateURL("/home/work/seen/../down/..")
	if err != nil || u == nil || u.String() != "file:///home/work" {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateURLMultiHttp(t *testing.T) {
	u, err := UltimateURL("http://server/home/work/seen/../down/..")
	if err != nil || u == nil || u.String() != "http://server/home/work" {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateURLWindows(t *testing.T) {
	u, err := UltimateURL("c:\\windows\\something", "nothing")

	// off windows the '\\' is an escape
	if runtime.GOOS == "windows" {
		if err != nil || u == nil || u.String() != "file:///c:/windows/something/nothing" {
			t.Error("unexpected error", err, u)
		}
	} else {
		if err != nil || u == nil {
			t.Error("unexpected error", err, u)
			return
		}
		p, err := URLToRelativePath(u)
		if err != nil || p != "c:\\windows\\something/nothing" {
			t.Error("unexpected error", err, u, p)
			return
		}
	}
}

func TestUltimateUNC(t *testing.T) {
	u, err := UltimateURL("//server/share/path", "99/../100", "101")
	if err != nil || u == nil || u.String() != "file://server/share/path/100/101" {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateUNCServerOnly(t *testing.T) {
	u, err := UltimateURL("//server")
	if err != nil || u == nil || u.String() != "file://server/." {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateUnix(t *testing.T) {
	u, err := UltimateURL("/root/url/bin/", "somthing")
	if err != nil || u == nil || u.String() != "file:///root/url/bin/somthing" {
		t.Error("unexpected error", err, u)
	}
}

func TestUltimateUnixHome(t *testing.T) {
	u, err := UltimateURL("/root/url/bin/", "~/.home/thing")
	if err != nil || u == nil {
		t.Error("unexpected error", err, u)
		return
	}
	p, err := URLToRelativePath(u, "~")
	if err != nil || p != ".home/thing" {
		t.Error("unexpected error", err, u, p)
		return
	}
}

func TestUltimateUnixHomeAlone(t *testing.T) {
	home, _ := homedir.Dir()
	home = filepath.ToSlash(home)
	u, err := UltimateURL("~")
	if err != nil || u == nil || !strings.HasSuffix(u.String(), home) {
		t.Error("unexpected error", err, u)
		return
	}

	p, err := URLToRelativePath(u, home)
	if err != nil || p != "." {
		t.Error("unexpected error", err, u, p)
		return
	}
}

func TestJoinEmptyAdd(t *testing.T) {
	res := join("root", "")
	if res != "root" {
		t.Error("join unexpected", res)
	}
}

func TestJoinEmptyRoot(t *testing.T) {
	res := join("", "add")
	if res != "add" {
		t.Error("join unexpected", res)
	}
}

func TestJoinTrailingSlash(t *testing.T) {
	res := join("root", "add/")
	if res != "root/add" {
		t.Error("join unexpected", res)
	}
}

func TestJoinTrailingRootSlash(t *testing.T) {
	res := join("root/", "add/")
	if res != "root//add" {
		t.Error("join unexpected", res)
	}
}

func TestURLToPathFromHttps(t *testing.T) {
	u, _ := url.Parse("https://server/data")

	_, err := URLToPath(u)
	if err == nil {
		t.Error("expected error")
	}
}

func TestURLToPathFromUnc(t *testing.T) {
	u, _ := url.Parse("file://server/data")

	p, err := URLToPath(u)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if p != filepath.FromSlash("//server/data") {
		t.Error("unexpected path", p)
	}
}

func TestURLToPathFromLocal(t *testing.T) {
	u, _ := url.Parse("file:///server/data")

	p, err := URLToPath(u)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if p != filepath.FromSlash("/server/data") {
		t.Error("unexpected path", p)
	}
}

func TestURLToRelativePathFromHttps(t *testing.T) {
	u, _ := url.Parse("https://server/data")

	_, err := URLToRelativePath(u, "/")
	if err == nil {
		t.Error("expected error")
	}
}

func TestURLToRelativePathFromLocal(t *testing.T) {
	u, _ := url.Parse("file:///data/local/bin")

	p, err := URLToRelativePath(u, "/data/local")
	if err != nil {
		t.Error("unexpected error", err)
	}

	if p != filepath.FromSlash("bin") {
		t.Error("unexpected path", p)
	}
}

func TestGetURLParentLocation(t *testing.T) {
	u, _ := url.Parse("file:///data/local/bin")
	cpy := GetURLParentLocation(u)

	if cpy == u {
		t.Error("not copy")
	}

	if cpy.String() != "file:///data/local" {
		t.Error("not cpy error", cpy)
	}
}

func TestUltimateURLMergePathsFile(t *testing.T) {
	location, err := UltimateURL("/root", "thing")

	if err != nil || location.String() != "file:///root/thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestUltimateURLMergePathsFileAbs(t *testing.T) {
	location, err := UltimateURL("/root", "/thing")

	if err != nil || location.String() != "file:///thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestUltimateURLMergePathsUrl(t *testing.T) {
	location, err := UltimateURL("https://root:9090/data", "thing")

	if err != nil || location.String() != "https://root:9090/data/thing" {
		t.Error("merge (1) url", location, err)
	}
}

func TestUltimateURLMergePathsUrlAbs(t *testing.T) {
	location, err := UltimateURL("https://root:9090/data", "/thing")

	if err != nil || location.String() != "https://root:9090/thing" {
		t.Error("merge (1) url", location, err)
	}
}

func TestUltimateURLMergePathsFileUrlAppend(t *testing.T) {
	location, err := UltimateURL("file:/root/data", "thing")

	if err != nil || location.String() != "file:///root/data/thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestUltimateURLMergePathsFileUrlAppendThreeSlash(t *testing.T) {
	location, err := UltimateURL("file:///root/data", "thing")

	if err != nil || location.String() != "file:///root/data/thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestUltimateURLMergePathsFileUrl(t *testing.T) {
	location, err := UltimateURL("file:/root/data", "/thing")

	if err != nil || location.String() != "file:///thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestUltimateURLMergePathsFileUrlThreeSlash(t *testing.T) {
	location, err := UltimateURL("file:///root/data", "/thing")

	if err != nil || location.String() != "file:///thing" {
		t.Error("merge (1) file", location, err)
	}
}

func TestGetBaseLocationWindows(t *testing.T) {
	loc, err := GetParentLocation("c:\\windows\\data")
	if runtime.GOOS == "windows" {
		if err != nil || loc.String() != "c:\\windows" {
			t.Error("parsing windows", loc, err)
		}
	} else {
		if err != nil || !strings.HasSuffix(loc.String(), "resource") ||
			!strings.HasPrefix(loc.String(), "file:///") {
			t.Error("parsing escaped path", loc, err)
		}
	}
}

func TestGetBaseLocationFile(t *testing.T) {
	loc, err := GetParentLocation("/root/data/file.txt")

	if err != nil || loc.String() != "file:///root/data" {
		t.Error("parsing file", loc, err)
	}
}

func TestGetBaseLocationFileUrl(t *testing.T) {
	loc, err := GetParentLocation("file:///root/data/file.txt")

	if err != nil || loc.String() != "file:///root/data" {
		t.Error("parsing file", loc, err)
	}
}

func TestGetBaseLocationWebUrl(t *testing.T) {
	loc, err := GetParentLocation("http://space/root/data/file.txt")

	if err != nil || loc.String() != "http://space/root/data" {
		t.Error("parsing file", loc, err)
	}
}

func TestRelativeHttp(t *testing.T) {
	u, err := url.Parse("http://space/root/data/file.txt")
	if err != nil {
		panic(err)
	}

	p := Relative(u)

	if p != "http://space/root/data/file.txt" {
		t.Error("unexpected", p)
	}
}

func TestRelativeLocal(t *testing.T) {
	u, err := UltimateURL("reader.go")
	if err != nil {
		panic(err)
	}

	p := Relative(u)

	if p != "reader.go" {
		t.Error("unexpected", p)
	}
}

func TestRelativeStringHttp(t *testing.T) {
	p := Relative("http://space/root/data/file.txt")

	if p != "http://space/root/data/file.txt" {
		t.Error("unexpected", p)
	}
}

func TestRelativeStringLocal(t *testing.T) {
	p := Relative("reader.go")

	if p != "reader.go" {
		t.Error("unexpected", p)
	}
}

func TestRelativePanicType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		}
	}()

	_ = Relative(10)
}
