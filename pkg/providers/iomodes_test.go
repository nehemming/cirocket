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
	"testing"
)

func modeString(mode IOMode) string {
	return fmt.Sprintf("%d:%06b", mode, mode)
}

func TestIOModes(t *testing.T) {
	r := modeString(IOModeNone)
	if r != "0:000000" {
		t.Error("IOModeNone", r)
	}

	r = modeString(IOModeInput)
	if r != "1:000001" {
		t.Error("IOModeInput", r)
	}

	r = modeString(IOModeOutput)
	if r != "2:000010" {
		t.Error("IOModeOutput", r)
	}

	r = modeString(IOModeError)
	if r != "4:000100" {
		t.Error("IOModeError", r)
	}

	r = modeString(IOModeTruncate)
	if r != "8:001000" {
		t.Error("IOModeCreate", r)
	}

	r = modeString(IOModeAppend)
	if r != "16:010000" {
		t.Error("IOModeCreate", r)
	}
}

func TestIOModesString(t *testing.T) {
	r := IOModeNone.String()
	if r != "-----" {
		t.Error("IOModeNone", r)
	}

	r = IOModeInput.String()
	if r != "i----" {
		t.Error("IOModeInput", r)
	}

	r = IOModeOutput.String()
	if r != "-o---" {
		t.Error("IOModeOutput", r)
	}
	r = IOModeError.String()
	if r != "--e--" {
		t.Error("IOModeError", r)
	}

	r = IOModeAppend.String()
	if r != "---a-" {
		t.Error("IOModeAppend", r)
	}

	r = IOModeTruncate.String()
	if r != "----t" {
		t.Error("IOModeTruncate", r)
	}
}
