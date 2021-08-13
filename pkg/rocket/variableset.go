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

import "sync"

type variableSet struct {
	variables map[string]string
	rwLock    sync.RWMutex
}

func newVariableSet() *variableSet {
	return &variableSet{
		variables: make(map[string]string),
	}
}

// All returns a copy of all the exported variables.
func (vs *variableSet) All() map[string]string {
	vs.rwLock.RLock()
	defer vs.rwLock.RUnlock()

	m := make(map[string]string)

	for k, v := range vs.variables {
		m[k] = v
	}

	return m
}

// Set assigns the variable to the set.
func (vs *variableSet) Set(key, value string) {
	vs.rwLock.Lock()
	defer vs.rwLock.Unlock()
	vs.variables[key] = value
}

// Get returns the variable value and a boolean to indicate if present.
func (vs *variableSet) Get(key string) (string, bool) {
	vs.rwLock.RLock()
	defer vs.rwLock.RUnlock()
	v, ok := vs.variables[key]
	return v, ok
}
