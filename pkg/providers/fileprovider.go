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
	"net/url"
	"os"

	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
)

type (
	// FileDetailer provides details on a file based resource provider.
	FileDetailer interface {

		// URL returns the url
		URL() *url.URL

		// FilePath returns the os path to the file
		FilePath() string

		// IOMode is th eio moe supported by the file
		IOMode() IOMode

		// File mode is the file's permission mask
		FileMode() os.FileMode

		// InMode returns true if the mode is supported by the file
		InMode(mode IOMode) bool
	}

	fileResourceProvider struct {
		url      *url.URL
		filePath string
		ioMode   IOMode
		fileMode os.FileMode
		optional bool
	}
)

// NewFileProvider creates a file provider.
func NewFileProvider(path string, ioMode IOMode, fileMode os.FileMode, optional bool) (ResourceProvider, error) {
	if path == "" {
		return nil, errors.New("path is blank")
	}

	if ioMode&(IOModeOutput|IOModeError) != IOModeNone {
		if ioMode&(IOModeTruncate|IOModeAppend) == (IOModeTruncate | IOModeAppend) {
			return nil, errors.New("both truncate and append have been specified, please select only one")
		} else if ioMode&(IOModeTruncate|IOModeAppend) == IOModeNone {
			return nil, errors.New("neither truncate nor append have been specified, please select only one")
		}
		// Supports output
	} else if ioMode&IOModeInput != IOModeInput {
		// not in or out
		return nil, errors.New("mode is neither input nor output")
	}

	// confirm is a file
	url, err := resource.UltimateURL(path)
	if err != nil {
		return nil, err
	}
	path, err = resource.URLToPath(url)
	if err != nil {
		return nil, err
	}

	return &fileResourceProvider{
		url:      url,
		filePath: path,
		ioMode:   ioMode,
		fileMode: fileMode,
		optional: optional,
	}, nil
}

func (fp *fileResourceProvider) FilePath() string {
	return fp.filePath
}

func (fp *fileResourceProvider) URL() *url.URL {
	u := *fp.url
	return &u
}

func (fp *fileResourceProvider) IOMode() IOMode {
	return fp.ioMode
}

func (fp *fileResourceProvider) FileMode() os.FileMode {
	return fp.fileMode
}

func (fp *fileResourceProvider) InMode(mode IOMode) bool {
	return fp.ioMode&mode == mode
}

func (fp *fileResourceProvider) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	if fp.ioMode&IOModeInput != IOModeInput {
		return nil, errors.New("input is not supported")
	}

	rc, err := resource.OpenRead(ctx, fp.url)
	if err != nil {
		if !fp.optional || resource.IsNotFoundError(err) == nil {
			return nil, err
		}

		rc = resource.NewReadCloser(new(bytes.Buffer))
	}
	return rc, nil
}

func (fp *fileResourceProvider) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	if fp.ioMode&(IOModeOutput|IOModeError) == IOModeNone {
		return nil, errors.New("output is not supported")
	}

	if (fp.ioMode & IOModeTruncate) == IOModeTruncate {
		return resource.OpenTruncate(ctx, fp.url, fp.fileMode)
	} else if (fp.ioMode & IOModeAppend) == IOModeAppend {
		return resource.OpenAppend(ctx, fp.url, fp.fileMode)
	} else {
		panic("validation should have caught missing append or truncate")
	}
}
