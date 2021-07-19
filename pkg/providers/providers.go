package providers

import (
	"context"
	"io"
	"sync"
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
)

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
