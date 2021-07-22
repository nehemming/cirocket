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

package loggee

import (
	"fmt"
	"strings"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func testMultiErrorFormatter(es []error) string {
	if len(es) == 1 {
		return es[0].Error()
	}

	text := make([]string, len(es))
	for i, err := range es {
		text[i] = fmt.Sprintf("%s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred: %s",
		len(es), strings.Join(text, "; "))
}

func TestMultipleErrorFormattingOnNil(t *testing.T) {
	SetMultiErrorFormatting(testMultiErrorFormatter)

	err := BindMultiErrorFormatting(nil)
	if err != nil {
		t.Error("Mismatch nil")
	}
}

func TestMultipleErrorFormattingSimple(t *testing.T) {
	SetMultiErrorFormatting(testMultiErrorFormatter)

	err := BindMultiErrorFormatting(errors.New("simple"))

	if err.Error() != "simple" {
		t.Error("Mismatch simple", err)
	}

	// err = multierror.Append(err, errors.Wrapf(e, "close[%d]", i))
}

func TestMultipleErrorFormattingSingle(t *testing.T) {
	SetMultiErrorFormatting(testMultiErrorFormatter)

	var err error
	err = multierror.Append(err, errors.New("one"))

	err = BindMultiErrorFormatting(err)

	if err.Error() != "one" {
		t.Error("Mismatch single", err)
	}
}

func TestMultipleErrorFormattingMultiple(t *testing.T) {
	SetMultiErrorFormatting(testMultiErrorFormatter)

	var err error
	err = multierror.Append(err, errors.New("one"))
	err = multierror.Append(err, errors.New("two"))

	err = BindMultiErrorFormatting(err)

	if err.Error() != "2 errors occurred: one; two" {
		t.Error("Mismatch multiple", err)
	}
}
