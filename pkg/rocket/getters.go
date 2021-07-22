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
	"os"
	"strings"
)

type (
	// Getter returns a value for an input key, iif the key is
	// not found an empty string is returned.
	Getter interface {
		// Get a value
		Get(key string) string

		// All returns all the entries
		All() map[string]string
	}

	// KeyValueGetter is a value map with fallback to a parent Getter.
	KeyValueGetter struct {
		kv     map[string]string
		parent Getter
	}
)

type osEnvGetter struct{}

// Gets an environment variable's value.
func (osEnvGetter) Get(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return ""
}

// All returns all environment variables is use.
func (osEnvGetter) All() map[string]string {
	m := make(map[string]string)

	for _, p := range os.Environ() {
		split := strings.SplitN(p, "=", 2)

		if len(split) > 1 {
			m[split[0]] = split[1]
		}
	}

	return m
}

// NewKeyValueGetter creates a new KeyValueGetter.
func NewKeyValueGetter(parent Getter) *KeyValueGetter {
	return &KeyValueGetter{
		kv:     make(map[string]string),
		parent: parent,
	}
}

// Get returns the value for a key.
func (kvg *KeyValueGetter) Get(key string) string {
	if v, ok := kvg.kv[key]; ok {
		return v
	}
	if kvg.parent != nil {
		return kvg.parent.Get(key)
	}
	return ""
}

// All returns the values tp down, where child values override parent key values.
func (kvg *KeyValueGetter) All() map[string]string {
	var m map[string]string

	if kvg.parent != nil {
		m = kvg.parent.All()
	} else {
		m = make(map[string]string)
	}

	for k, v := range kvg.kv {
		m[k] = v
	}

	return m
}
