package rocket

import (
	"os"
	"strings"
)

type (
	// Getter returns a value for an input key, iif the key is
	// not found an empty string is returned
	Getter interface {
		// Get a value
		Get(key string) string

		// All returns all the entries
		All() map[string]string
	}

	// KeyValueGetter is a value map with fallback to a parent Getter
	KeyValueGetter struct {
		kv     map[string]string
		parent Getter
	}
)

type osEnvGetter struct{}

// Gets an environment variable's value
func (osEnvGetter) Get(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return ""
}

// All returns all environment variables is use
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

// NewKeyValueGetter creates a new KeyValueGetter
func NewKeyValueGetter(parent Getter) *KeyValueGetter {
	return &KeyValueGetter{
		kv:     make(map[string]string),
		parent: parent,
	}
}

// Get returns the value for a key
func (kvg *KeyValueGetter) Get(key string) string {
	if v, ok := kvg.kv[key]; ok {
		return v
	}
	if kvg.parent != nil {
		return kvg.parent.Get(key)
	}
	return ""
}

// All returns the values tp down, where child values override parent key values
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
