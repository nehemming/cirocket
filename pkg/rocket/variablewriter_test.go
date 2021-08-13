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
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestNewVariableWriter(t *testing.T) {
	capComm := NewCapComm("here", stdlog.New())

	vw := newVariableWriter(capComm, "test")

	if len(vw.data.Bytes()) != 0 || vw.name != "test" || vw.capComm != capComm {
		t.Error("unexpected values", vw.data, vw.capComm, vw.name)
	}
}

func TestOpenRead(t *testing.T) {
	capComm := NewCapComm("here", stdlog.New())

	vw := newVariableWriter(capComm, "test")

	if _, err := vw.OpenRead(context.Background()); err == nil {
		t.Error("no error for open variable reader")
	}
}

func TestOpenWrite(t *testing.T) {
	capComm := NewCapComm("here", stdlog.New())

	vw := newVariableWriter(capComm, "test")

	w, err := vw.OpenWrite(context.Background())
	if err != nil {
		t.Error("unexpected", err)
	}

	_, err = w.Write([]byte("hello"))
	if err != nil {
		t.Error("unexpected", err)
	}

	if vw.data.String() != "hello" {
		t.Error("mismatch", vw.data.String())
	}

	v, _ := capComm.exportTo.Get("test")

	if v != "" {
		t.Error("too soon")
	}

	vw.Close()

	exp := capComm.exportTo.All()

	if exp["test"] != "hello" {
		t.Error("mismatch", vw.data.String())
	}
}
