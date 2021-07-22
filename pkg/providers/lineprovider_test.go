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

package providers

import (
	"context"
	"testing"
)

func TestNewLineProvider(t *testing.T) {
	called := new(bool)
	*called = false

	fn := func(s string) {
		*called = true
		if s != "hello" {
			t.Error("expected hello", s)
		}
	}

	provider := NewLineProvider(fn)

	// Check read fails
	_, err := provider.OpenRead(context.Background())
	if err == nil {
		t.Error("error expected")
	}

	writer, err := provider.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
		writer.Close()
		return
	}

	writer.Close()

	if !*called {
		t.Error("not called fn")
	}
}
