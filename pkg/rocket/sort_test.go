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
	"sort"
	"testing"
)

func TestTaskTypeInfoListSort(t *testing.T) {
	data := TaskTypeInfoList{{"a", "a desc"}, {"c", "c desc"}, {"b", "a desc"}}

	if data.Len() != 3 {
		t.Error("len", data.Len())
	}

	if data.Less(1, 2) {
		t.Error("less reversed")
	}

	sort.Sort(data)

	if data[0].Type != "a" || data[1].Type != "b" {
		t.Error("sorting issue", data)
	}
}
