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
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/pkg/errors"
)

// CloserList is a list of io.Closers.
type CloserList struct {
	closers []io.Closer
}

// NewCloserList creates a new closer list.
func NewCloserList(items ...interface{}) *CloserList {
	return new(CloserList).Append(items...)
}

// Append a closer or slice of closers to the closerlist.
func (cl *CloserList) Append(items ...interface{}) *CloserList {
	for _, item := range items {
		switch v := item.(type) {
		case io.Closer:
			cl.closers = append(cl.closers, v)
		case []io.Closer:
			cl.closers = append(cl.closers, v...)
		}
	}

	return cl
}

// Close closes all closers in the list returning all errors as a single error.
func (cl *CloserList) Close() error {
	var err error

	for i, c := range cl.closers {
		e := c.Close()

		if e != nil {
			err = multierror.Append(err, errors.Wrapf(e, "close[%d]", i))
		}
	}

	return loggee.BindMultiErrorFormatting(err)
}
