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
	"context"
	"io"
	"sync"

	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
)

type (
	// IOMode indicates the modes of operation the file detail supports.
	IOMode uint32

	// ResourceID is a the ID of a resource.
	ResourceID string

	// ResourceProvider creates resources.
	// Implementers must implement both functions but can return an error
	// for un supported requests.  For example a Web Getter cannot be opened for writes so should return an error.
	// All resources returned without an error should expect Close to be called on the ReadCloser or WriteCloser.
	ResourceProvider interface {
		OpenRead(ctx context.Context) (io.ReadCloser, error)
		OpenWrite(ctx context.Context) (io.WriteCloser, error)
	}

	// ResourceProviderMap is a map collection of resource prooviders.
	ResourceProviderMap map[ResourceID]ResourceProvider

	idempotentCloser struct {
		once   sync.Once
		closer io.Closer
	}

	nopReaderCloser struct {
		reader io.Reader
	}

	nopWriterCloser struct {
		writer io.Writer
	}
)

// Copy creates a copy of a resource map.
func (m ResourceProviderMap) Copy() ResourceProviderMap {
	copy := make(ResourceProviderMap)

	for k, v := range m {
		copy[k] = v
	}

	return copy
}

// NewIdempotentCloser creates a closer that calls the underlying closer once only.
func NewIdempotentCloser(closer io.Closer) io.Closer {
	return &idempotentCloser{
		closer: closer,
	}
}

// Close the resource and return an error.
func (idem *idempotentCloser) Close() error {
	var err error

	// calling close once, on that occasion err will be set, other calls err will
	// remain nil as not set.
	idem.once.Do(func() { err = idem.closer.Close() })

	return err
}

// NewNonClosingReaderProvider attaches an existing reader (i.e. stdin) to a provider.
func NewNonClosingReaderProvider(reader io.Reader) ResourceProvider {
	return &nopReaderCloser{
		reader: reader,
	}
}

// NewNonClosingWriterProvider attaches an existing writer (i.e. stdout) to a provider.
func NewNonClosingWriterProvider(writer io.Writer) ResourceProvider {
	return &nopWriterCloser{
		writer: writer,
	}
}

func (rc *nopReaderCloser) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return nil, errors.New("output is not supported")
}

func (rc *nopReaderCloser) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return resource.NewReadCloser(rc.reader), nil
}

func (wc *nopWriterCloser) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return resource.NewWriteCloser(wc.writer), nil
}

func (wc *nopWriterCloser) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return nil, errors.New("input is not supported")
}
