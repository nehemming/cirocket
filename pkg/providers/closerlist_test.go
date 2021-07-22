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
	"fmt"
	"io"
	"testing"
)

type testCloser struct {
	counter int
}

func (tc *testCloser) Close() error {
	tc.counter++

	return nil
}

type testErrorCloser struct {
	id int
}

func (tc *testErrorCloser) Close() error {
	return fmt.Errorf("err %d", tc.id)
}

func TestCloserListNewEmpty(t *testing.T) {
	cl := NewCloserList()

	if len(cl.closers) != 0 {
		t.Error("NewCloserList is not empty", len(cl.closers))
	}
}

func TestCloserListNewSingleCloser(t *testing.T) {
	c := &testCloser{}

	cl := NewCloserList(c)

	if len(cl.closers) != 1 {
		t.Error("NewCloserList is not 1", len(cl.closers))
	}

	cl.Close()

	if c.counter != 1 {
		t.Error("CloserList did not close")
	}
}

func TestCloserListNewTwoCloser(t *testing.T) {
	c := &testCloser{}
	c1 := &testCloser{}

	cl := NewCloserList(c, c1)

	if len(cl.closers) != 2 {
		t.Error("NewCloserList is not 2", len(cl.closers))
	}

	cl.Close()

	if c.counter != 1 {
		t.Error("CloserList did not close c")
	}

	if c1.counter != 1 {
		t.Error("CloserList did not close c1")
	}
}

func TestCloserListNewMultipleClosers(t *testing.T) {
	c := make([]io.Closer, 5)
	for i := 0; i < len(c); i++ {
		c[i] = &testCloser{}
	}

	cl := NewCloserList(c)

	if len(cl.closers) != len(c) {
		t.Error("NewCloserList is not", len(c), len(cl.closers))
	}

	cl.Close()

	for i := 0; i < len(c); i++ {
		if c[i].(*testCloser).counter != 1 {
			t.Error("CloserList did not close", i)
		}
	}
}

func TestCloserListNewMultipleDeepClosers(t *testing.T) {
	c := make([]io.Closer, 5)
	for i := 0; i < len(c); i++ {
		c[i] = &testCloser{}
	}

	cl := NewCloserList(c)
	cSingle := &testCloser{}

	cl2 := NewCloserList(cl, cSingle)

	if len(cl2.closers) != 2 {
		t.Error("NewCloserList is shallow", len(cl.closers))
	}

	cl2.Close()

	for i := 0; i < len(c); i++ {
		if c[i].(*testCloser).counter != 1 {
			t.Error("CloserList did not close", i)
		}
	}

	if cSingle.counter != 1 {
		t.Error("CloserList did not close cSingle")
	}
}

func TestCloserListAppendSingleCloser(t *testing.T) {
	cl := NewCloserList()

	c := &testCloser{}

	cl.Append(c)

	if len(cl.closers) != 1 {
		t.Error("NewCloserList is not 1", len(cl.closers))
	}

	cl.Close()

	if c.counter != 1 {
		t.Error("CloserList did not close")
	}
}

func TestCloserListAppendTwoCloser(t *testing.T) {
	cl := NewCloserList()

	c := &testCloser{}
	c1 := &testCloser{}

	cl.Append(c, c1)

	if len(cl.closers) != 2 {
		t.Error("NewCloserList is not 2", len(cl.closers))
	}

	err := cl.Close()
	if err != nil {
		t.Error("close", err)
	}

	if c.counter != 1 {
		t.Error("CloserList did not close c")
	}

	if c1.counter != 1 {
		t.Error("CloserList did not close c1")
	}
}

func TestCloserListAppendMultipleClosers(t *testing.T) {
	cl := NewCloserList()

	c := make([]io.Closer, 5)
	for i := 0; i < len(c); i++ {
		c[i] = &testCloser{}
	}

	cl.Append(c)

	if len(cl.closers) != len(c) {
		t.Error("NewCloserList is not", len(c), len(cl.closers))
	}

	err := cl.Close()
	if err != nil {
		t.Error("close", err)
	}

	for i := 0; i < len(c); i++ {
		if c[i].(*testCloser).counter != 1 {
			t.Error("CloserList did not close", i)
		}
	}
}

func TestCloserLisAppendMultipleDeepClosers(t *testing.T) {
	cl := NewCloserList()

	c := make([]io.Closer, 5)
	for i := 0; i < len(c); i++ {
		c[i] = &testCloser{}
	}

	cl.Append(c)
	cSingle := &testCloser{}

	cl2 := NewCloserList()

	cl2.Append(cl, cSingle)

	if len(cl2.closers) != 2 {
		t.Error("NewCloserList is shallow", len(cl.closers))
	}

	err := cl2.Close()
	if err != nil {
		t.Error("close", err)
	}
	for i := 0; i < len(c); i++ {
		if c[i].(*testCloser).counter != 1 {
			t.Error("CloserList did not close", i)
		}
	}

	if cSingle.counter != 1 {
		t.Error("CloserList did not close cSingle")
	}
}

func TestCloserListNewSingleCloserErrors(t *testing.T) {
	c := &testErrorCloser{}

	cl := NewCloserList(c)

	if len(cl.closers) != 1 {
		t.Error("NewCloserList is not 1", len(cl.closers))
	}

	err := cl.Close()
	if err == nil {
		t.Error("No error on close")
	}
}

func TestCloserListNewMultipleCloserErrors(t *testing.T) {
	c := &testErrorCloser{}

	cl := NewCloserList(c, &testErrorCloser{1})

	if len(cl.closers) != 2 {
		t.Error("NewCloserList is not 2", len(cl.closers))
	}

	err := cl.Close()
	if err == nil {
		t.Error("No error on close")
	}
}
