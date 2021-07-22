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
	"context"
	"io"
	"runtime"
	"strings"
	"testing"
)

func TestReadResourceZero(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx)
	if err != nil || len(b) != 0 {
		t.Error("unexpected error", err, b)
	}
}

func TestReadResourceBadProtocol(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx, "scp://root@host:/root/ids/rules")
	if err == nil || b != nil {
		t.Error("unexpected error", err, b)
	}
}

func TestReadResourceUncBadProtocol(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	ctx := context.Background()
	b, err := ReadResource(ctx, "//legacy/data/nothere")
	if err == nil || b != nil {
		t.Error("unexpected error", err, b)
	}
}

func TestReadResourceLocalFile(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx, "search.go")
	if err != nil || !strings.HasPrefix(string(b), "/*") {
		t.Error("unexpected", err, string(b))
	}
}

func TestReadResourceLocalTestFile(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx, "./testdata", "one.txt")
	if err != nil || !strings.HasPrefix(string(b), "Hello") {
		t.Error("unexpected", err, string(b))
	}
}

func TestReadResourceRemoteWeb(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx, "https://raw.githubusercontent.com/nehemming/cirocket/master/README.md")
	if err != nil || !strings.HasPrefix(string(b), "#") {
		t.Error("unexpected", err, string(b))
	}
}

func TestReadResourceRemoteWebNotFound(t *testing.T) {
	ctx := context.Background()
	_, err := ReadResource(ctx, "https://raw.githubusercontent.com/nehemming/cirocket/master/notREADME.md")
	if err == nil {
		t.Error("no error")
	}
	if nfe, ok := err.(*NotFoundError); !ok {
		t.Error("not a not found error", err)
	} else if nfe.resource != "https://raw.githubusercontent.com/nehemming/cirocket/master/notREADME.md" {
		t.Error("not resource", nfe.resource)
	}
}

func TestReadResourceRemoteFileSplit(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx,
		"https://raw.githubusercontent.com/nehemming",
		"cirocket/master/README.md")
	if err != nil || !strings.HasPrefix(string(b), "#") {
		t.Error("unexpected", err, string(b))
	}
}

func TestReadResourceRemoteFileSplitMltioots(t *testing.T) {
	ctx := context.Background()
	b, err := ReadResource(ctx,
		"c:\\windows", "//www/data/me",
		"https://raw.githubusercontent.com/nehemming",
		"cirocket/master/README.md")
	if err != nil || !strings.HasPrefix(string(b), "#") {
		t.Error("unexpected", err, string(b))
	}
}

func TestOpenReadFile(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("./testdata", "one.txt")
	if err != nil {
		t.Error("unexpected", err)
	}

	f, err := OpenRead(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil || !strings.HasPrefix(string(b), "Hello") {
		t.Error("unexpected", err, string(b))
	}
}

func TestOpenReadNotFoundErrors(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("./testdata", "none.txt")
	if err != nil {
		t.Error("unexpected", err)
	}

	_, err = OpenRead(ctx, u)
	if err == nil {
		t.Error("expected error")
	}

	if IsNotFoundError(err) == nil {
		t.Error("expected not found error", err)
	}
}

func TestOpenReadURLErrors(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("https://raw.githubusercontent.com/nehemming",
		"cirocket/master/README.md")
	if err != nil {
		t.Error("unexpected", err)
	}

	f, err := OpenRead(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil || !strings.HasPrefix(string(b), "#") {
		t.Error("unexpected", err, string(b))
	}
}

func TestOpenReadSchemeErrors(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("sftp://raw.githubusercontent.com/nehemming",
		"cirocket/master/README.md")
	if err != nil {
		t.Error("unexpected", err)
	}

	_, err = OpenRead(ctx, u)
	if err == nil {
		t.Error("expected error")
	}
}
