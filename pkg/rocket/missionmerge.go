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

func mergeMissions(mission, addition *Mission) {
	if mission.Name == "" {
		mission.Name = addition.Name
	}
	if mission.Version == "" {
		mission.Version = addition.Version
	}

	missionMergeEnv(mission, addition)

	if len(addition.Params) > 0 {
		missionMergeParams(mission, addition)
	}

	if len(addition.Stages) > 0 {
		missionMergeStages(mission, addition)
	}

	if len(addition.Sequences) > 0 {
		missionMergeSequences(mission, addition)
	}
}

func missionMergeEnv(mission, addition *Mission) {
	if len(addition.BasicEnv) > 0 {
		if mission.BasicEnv == nil {
			mission.BasicEnv = make(EnvMap)
		}
		for k, v := range addition.BasicEnv {
			if _, ok := mission.BasicEnv[k]; !ok {
				mission.BasicEnv[k] = v
			}
		}
	}
	if len(addition.Env) > 0 {
		if mission.Env == nil {
			mission.Env = make(EnvMap)
		}
		for k, v := range addition.Env {
			if _, ok := mission.Env[k]; !ok {
				mission.Env[k] = v
			}
		}
	}
}

func missionMergeParams(mission, addition *Mission) {
	m := make(map[string]bool)
	for _, p := range mission.Params {
		m[p.Name] = true
	}

	params := make([]Param, 0, len(addition.Params))
	for _, p := range addition.Params {
		if _, ok := m[p.Name]; !ok || p.Name == "" {
			params = append(params, p)
		}
	}

	mission.Params = append(mission.Params, params...)
}

func missionMergeStages(mission, addition *Mission) {
	m := make(map[string]bool)
	for _, st := range mission.Stages {
		m[st.Name] = true
	}

	stages := make([]Stage, 0, len(addition.Stages))
	for _, st := range addition.Stages {
		if _, ok := m[st.Name]; !ok || st.Name == "" {
			stages = append(stages, st)
		}
	}

	mission.Stages = append(mission.Stages, stages...)
}

func missionMergeSequences(mission, addition *Mission) {
	if mission.Sequences == nil {
		mission.Sequences = make(map[string][]string)
	}

	for k, seq := range addition.Sequences {
		if _, ok := mission.Sequences[k]; !ok {
			mission.Sequences[k] = seq
		}
	}
}
