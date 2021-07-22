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

func TestMissionMergeParams(t *testing.T) {
	// load them in

	mission := new(Mission)

	addition := &Mission{
		Name:    "hello",
		Version: "v20",
		Params: []Param{
			{Name: "one", Value: "1"},
		},
	}

	missionMergeParams(mission, addition)

	if len(mission.Params) != 1 {
		t.Error("missing addition", len(mission.Params))
	}

	if mission.Params[0].Name != "one" || mission.Params[0].Value != "1" {
		t.Error("mission.Params[0]", mission.Params[0])
	}
}

func TestMissionMergeParamsMany(t *testing.T) {
	// load them in

	mission := &Mission{
		Params: []Param{
			{Name: "one", Value: "0"},
			{Name: "two", Value: "2"},
		},
	}

	addition := &Mission{
		Params: []Param{
			{Name: "one", Value: "1"},
			{Name: "three", Value: "3"},
		},
	}

	missionMergeParams(mission, addition)

	if len(mission.Params) != 3 {
		t.Error("missing addition", len(mission.Params))
	}

	if mission.Params[0].Name != "one" || mission.Params[0].Value != "0" {
		t.Error("mission.Params[0]", mission.Params[0])
	}
	if mission.Params[1].Name != "two" || mission.Params[1].Value != "2" {
		t.Error("mission.Params[1]", mission.Params[1])
	}
	if mission.Params[2].Name != "three" || mission.Params[2].Value != "3" {
		t.Error("mission.Params[2]", mission.Params[2])
	}
}

func TestMergeMissions(t *testing.T) {
	mission := &Mission{}
	addition := &Mission{
		Name:    "add",
		Version: "1000.0.0",
		Params: []Param{
			{Name: "test", Value: "vone"},
		},
	}

	mergeMissions(mission, addition)

	if mission.Name != addition.Name {
		t.Error("missing name", mission.Name, addition.Name)
	}
	if mission.Version != addition.Version {
		t.Error("missing Version", mission.Version, addition.Version)
	}

	if len(mission.Params) != len(addition.Params) {
		t.Error("missing Params", mission.Params, addition.Params)
	}
}
