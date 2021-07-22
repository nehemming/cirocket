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
	multierror "github.com/hashicorp/go-multierror"
)

// SetMultiErrorFormatting allows the multiple error formatting to be globally set.
func SetMultiErrorFormatting(f func([]error) string) {
	multiErrorFormatter = f
}

var multiErrorFormatter func([]error) string

// BindMultiErrorFormatting is a helper function to attach a standard formatter
// to opt in hashicorp/go-multierror errors.
func BindMultiErrorFormatting(err error) error {
	if err != nil && multiErrorFormatter != nil {
		if multi, ok := err.(*multierror.Error); ok {
			multi.ErrorFormat = multiErrorFormatter
		}
	}

	return err
}
