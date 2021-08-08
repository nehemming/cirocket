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

import "testing"

func TestLocationMoreThanOneSpecified(t *testing.T) {
	inc := Location{
		Path:   "1234",
		Inline: "5678",
	}

	err := inc.Validate()

	if err == nil || err.Error() != "more than one source was specified, only one is permitted" {
		t.Error("unexpected", err)
	}
}

func TestLocationNoSources(t *testing.T) {
	inc := Location{}

	err := inc.Validate()

	if err == nil || err.Error() != "no source was specified" {
		t.Error("unexpected", err)
	}
}

func TestLocationPathSources(t *testing.T) {
	inc := Location{Path: "12"}

	err := inc.Validate()
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestLocationInlineSources(t *testing.T) {
	inc := Location{Inline: "12"}

	err := inc.Validate()
	if err != nil {
		t.Error("unexpected", err)
	}
}
