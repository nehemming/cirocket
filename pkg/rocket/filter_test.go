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
	"runtime"
	"testing"
)

func TestIsFilteredInclude(t *testing.T) {
	var f *Filter
	if f.IsFiltered() != false {
		t.Error("Nil should not filter")
	}

	// Included testing
	f = new(Filter)
	if f.IsFiltered() != false {
		t.Error("Empty should not filter")
	}
	f.IncludeOS = []string{runtime.GOOS}
	if f.IsFiltered() != false {
		t.Error("Same Os should not filter")
	}
	f.IncludeArch = []string{runtime.GOARCH}
	if f.IsFiltered() != false {
		t.Error("Same Arch should not filter")
	}
}

func TestIsFilteredSkip(t *testing.T) {
	f := new(Filter)
	f.Skip = true
	if f.IsFiltered() != true {
		t.Error("Skip true should filter")
	}
}

func TestIsFilteredNotInclude(t *testing.T) {
	var f *Filter
	if f.IsFiltered() != false {
		t.Error("Nil should not filter")
	}

	// Not included testing
	f = new(Filter)
	f.IncludeOS = []string{runtime.GOOS + "nope"}
	if f.IsFiltered() != true {
		t.Error("Diff Os should filter")
	}
	f.IncludeArch = []string{runtime.GOARCH + "nope"}
	if f.IsFiltered() != true {
		t.Error("Diff Arch should filter")
	}
}

func TestIsFilteredExclude(t *testing.T) {
	var f *Filter

	if f.IsFiltered() != false {
		t.Error("Nil should not filter")
	}

	// Exclude
	f = new(Filter)
	f.ExcludeOS = []string{runtime.GOOS}
	if f.IsFiltered() != true {
		t.Error("Same Os should exclude filter")
	}
	f = &Filter{}
	f.ExcludeArch = []string{runtime.GOARCH}
	if f.IsFiltered() != true {
		t.Error("Same Arch should exclude filter")
	}
}

func TestIsFilteredNotExclude(t *testing.T) {
	var f *Filter

	if f.IsFiltered() != false {
		t.Error("Nil should not filter")
	}

	// Non exclude test
	f = new(Filter)
	f.ExcludeOS = []string{runtime.GOOS + "nope"}
	if f.IsFiltered() != false {
		t.Error("Diff Os should exclude filter")
	}
	f = &Filter{}
	f.ExcludeArch = []string{runtime.GOARCH + "nope"}
	if f.IsFiltered() != false {
		t.Error("Diff Arch should exclude filter")
	}
}
