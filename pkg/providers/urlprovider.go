package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
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

	req, err := http.NewRequest("GET", rp.url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "creating request for %s", rp.url)
	}

	resp, err := ctxhttp.Do(ctxTimeout, nil, req)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s", rp.url)
	}
	defer resp.Body.Close()

	b := new(bytes.Buffer)
	if !rp.optional || resp.StatusCode != http.StatusNotFound {
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			// Bad response
			return nil, fmt.Errorf("response (%d) %s for %s", resp.StatusCode, resp.Status, rp.url)
		}

		// make a copy so we can close the response body here, cannot escape the ctxTimeout context
		_, err = io.Copy(b, resp.Body)

		if err != nil {
			return nil, errors.Wrapf(err, "extracting body %s", rp.url)
		}
	}

	// Return the body
	return &nopReaderCloser{b}, nil
}
