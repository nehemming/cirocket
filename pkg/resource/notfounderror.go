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
	"fmt"
	"strings"
)

// NotFoundError indicates a resource cannot be found.
type NotFoundError struct {
	cause    error
	resource string
}

// NewNotFoundError creates a new not found resource error.
// The cause should any error raised by the underlying resource open call.
// If no cause is provided a generic cause is added.
// The resource can be access via the Resource() method.
func NewNotFoundError(resource string, cause error) *NotFoundError {
	if cause == nil {
		cause = fmt.Errorf("not found")
	}
	return &NotFoundError{
		cause:    cause,
		resource: resource,
	}
}

// Resource returns the resource that was not found.
func (nfe *NotFoundError) Resource() string {
	return nfe.resource
}

// Error converts the errorinto a string.
func (nfe *NotFoundError) Error() string {
	ce := nfe.cause.Error()
	if strings.Contains(ce, nfe.resource) {
		return ce
	}
	return fmt.Sprintf("%s %s", nfe.resource, nfe.cause)
}

// Unwrap provides access to the causing error.
func (nfe *NotFoundError) Unwrap() error {
	return nfe.cause
}

// IsNotFoundError checks if an error is a *NotFoundError and if soo return it. If it is not nil is returned.
func IsNotFoundError(err error) *NotFoundError {
	if nfe, ok := err.(*NotFoundError); ok {
		return nfe
	}

	return nil
}
