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
	"bytes"
	"context"
	"io"
	"time"

	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
)

type (
	urlResourceProvider struct {
		url      string
		timeout  time.Duration
		optional bool
	}
)

// NewURLProvider creates a url provider.
func NewURLProvider(url string, timeout time.Duration, optional bool) (ResourceProvider, error) {
	if url == "" {
		return nil, errors.New("url is blank")
	}
	if timeout == 0 {
		timeout = time.Second * 30
	}

	return &urlResourceProvider{
		url:      url,
		timeout:  timeout,
		optional: optional,
	}, nil
}

func (rp *urlResourceProvider) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return nil, errors.New("output is not supported")
}

func (rp *urlResourceProvider) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, rp.timeout)
	defer cancel()

	b, err := resource.ReadResource(ctxTimeout, rp.url)
	if err != nil && (resource.IsNotFoundError(err) == nil || !rp.optional) {
		return nil, err
	}

	// Return the body (b can be safely nil)
	return resource.NewReadCloser(bytes.NewBuffer(b)), nil
}
