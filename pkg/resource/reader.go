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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
)

// ReadResource reads a resource into a byte array or returns an error.
// The resource parts are merged left to right using UltimateURL.
// Once defined the url is then access and read.
// Relative url paths are supported only if they ultimately resolve to an absolute url.
func ReadResource(ctx context.Context, pathParts ...string) ([]byte, error) {
	// Easy case, nothing request
	if len(pathParts) == 0 {
		return make([]byte, 0), nil
	}

	url, err := UltimateURL(pathParts...)
	if err != nil {
		return nil, err
	}

	return ReadURL(ctx, url)
}

// OpenRead opens a URL resource for reading.
//
// http(s) and file schemes are supported.
func OpenRead(ctx context.Context, url *url.URL) (io.ReadCloser, error) {
	switch url.Scheme {
	case "http", "https":
		b, err := readHTTP(ctx, url.String())
		if err != nil {
			return nil, err
		}
		return NewReadCloser(bytes.NewBuffer(b)), nil
	case "file":
		osPath, err := URLToPath(url)
		if err != nil {
			return nil, err
		}

		rc, err := os.Open(osPath)
		if err != nil {
			// capture not found
			if os.IsNotExist(err) {
				return nil, NewNotFoundError(osPath, err)
			}
			return nil, err
		}
		return rc, nil
	default:
		return nil, fmt.Errorf("scheme %s is not supported", url.Scheme)
	}
}

// ReadURL reads the contents located by the url into a byte slice or returns an error.
func ReadURL(ctx context.Context, url *url.URL) ([]byte, error) {
	switch url.Scheme {
	case "http", "https":
		return readHTTP(ctx, url.String())
	case "file":
		osPath, err := URLToPath(url)
		if err != nil {
			return nil, err
		}
		return readFile(osPath)
	default:
		return nil, fmt.Errorf("scheme %s is not supported", url.Scheme)
	}
}

func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		// capture not found
		if os.IsNotExist(err) {
			return nil, NewNotFoundError(path, err)
		}
		return nil, err
	}

	return b, nil
}

func readHTTP(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "creating request %s", url)
	}

	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewNotFoundError(url, nil)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Bad response
		return nil, fmt.Errorf("response (%d) %s for %s", resp.StatusCode, resp.Status, url)
	}

	b := new(bytes.Buffer)
	_, err = io.Copy(b, resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "extracting body %s", url)
	}

	return b.Bytes(), nil
}
