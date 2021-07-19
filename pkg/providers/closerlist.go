package providers

import (
	"io"

	"github.com/hashicorp/go-multierror"
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

	return bindMultiErrorFormatting(err)
}
