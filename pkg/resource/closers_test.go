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

package resource

import "testing"

type testCloser struct{}

func (tc *testCloser) Read(p []byte) (n int, err error) {
	return 1, nil
}

func (tc *testCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestNewReadCloser(t *testing.T) {
	tc := &testCloser{}

	rc := NewReadCloser(tc)

	data := make([]byte, 10)

	n, err := rc.Read(data)
	if n != 1 || err != nil {
		t.Error("unexpected", err, n)
	}

	err = rc.Close()
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestNewWriteCloser(t *testing.T) {
	tc := &testCloser{}

	rc := NewWriteCloser(tc)

	data := make([]byte, 10)

	n, err := rc.Write(data)
	if n != len(data) || err != nil {
		t.Error("unexpected", err, n)
	}

	err = rc.Close()
	if err != nil {
		t.Error("unexpected", err)
	}
}
