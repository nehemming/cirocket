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
	"net/url"
	"testing"
)

// Helper functions used for testing.

// isWindows lifted from resources, check for a windows volume.
func isWindows(part string) bool {
	if len(part) > 2 && part[1] == ':' {
		return true
	}
	return false
}

func fixUpWindows(url *url.URL) string {
	p := url.Path
	// in windows we get file:///c:/ so skip /c: /
	if len(p) > 3 && isWindows(p[1:]) {
		url.Path = p[3:]
	}
	return url.String()
}

func TestFixUpWindowsPath(t *testing.T) {
	u, _ := url.Parse("file:///c:/test")

	r := fixUpWindows(u)

	if r != "file:///test" {
		t.Error("unexpected url", u)
	}
}

func TestFixUpWindowsRoot(t *testing.T) {
	u, _ := url.Parse("file:///c:/")

	r := fixUpWindows(u)

	if r != "file:///" {
		t.Error("unexpected url", u)
	}
}

func TestFixUpLinux(t *testing.T) {
	u, _ := url.Parse("file:///root")

	r := fixUpWindows(u)

	if r != "file:///root" {
		t.Error("unexpected url", u)
	}
}
