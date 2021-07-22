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

import (
	"errors"
	"testing"
)

func TestNewNotFoundError(t *testing.T) {
	var err error //nolint

	err = NewNotFoundError("locA", nil)

	if err.Error() != "locA not found" {
		t.Error("unexpected string", err.Error())
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("not a *NotFoundError %T", err)
		return
	}
	nfe := err.(*NotFoundError)

	if nfe.Resource() != "locA" {
		t.Error("Location error", nfe.Resource())
	}

	cause := nfe.Unwrap()
	if cause == nil || cause.Error() != "not found" {
		t.Error("cause", cause)
	}
}

func TestNewNotFoundErrorWithCause(t *testing.T) {
	err := errors.New("its not there")

	err = NewNotFoundError("locA", err)

	if err.Error() != "locA its not there" {
		t.Error("unexpected string", err.Error())
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("not a *NotFoundError %T", err)
		return
	}
	nfe := err.(*NotFoundError)

	if nfe.Resource() != "locA" {
		t.Error("Location error", nfe.Resource())
	}

	cause := nfe.Unwrap()
	if cause == nil || cause.Error() != "its not there" {
		t.Error("cause", cause)
	}
}

func TestIsNotFoundError(t *testing.T) {
	err := errors.New("its not there")

	err = NewNotFoundError("locA", err)

	nfe := IsNotFoundError(err)
	if nfe == nil {
		t.Error("unexpected nil")
	}
}

func TestIsNotFoundErrorIsNil(t *testing.T) {
	err := errors.New("its not there")

	nfe := IsNotFoundError(err)
	if nfe != nil {
		t.Error("expected nil")
	}
}
