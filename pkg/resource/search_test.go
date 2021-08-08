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
	"net/url"
	"strings"
	"testing"
)

func TestSearchEmptySources(t *testing.T) {
	ctx := context.Background()

	_, _, err := Search(ctx, "something", nil)

	if err == nil {
		t.Error("Expected error")
	}
}

func TestSearchEmptyRel(t *testing.T) {
	ctx := context.Background()

	_, _, err := Search(ctx, "", nil, "one", "two")

	if err == nil {
		t.Error("Expected error")
	}
}

func TestSearchEmptyBadMerge(t *testing.T) {
	ctx := context.Background()

	_, _, err := Search(ctx, "",
		func(source string, u *url.URL, e *NotFoundError) {
			t.Error("progress!!")
		}, "", "two")

	if err == nil {
		t.Error("Expected error")
	}
}

func TestSearchFound(t *testing.T) {
	ctx := context.Background()

	b, u, err := Search(ctx, "reader.go", nil, ".")
	if err != nil || len(b) == 0 || strings.HasSuffix(u.Path, "resources/reader.go") {
		t.Error("unexpected error", err, u)
	}
}
