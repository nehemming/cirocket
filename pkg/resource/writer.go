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
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"

	"github.com/pkg/errors"
)

// IsWriteSupported returns true if the url can be written to by this package
// If true this does not mean the caller has permission r right too write only that
// Open/Write methods can be called.
func IsWriteSupported(url *url.URL) bool {
	switch url.Scheme {
	case "file":
		return url.Host == "" || runtime.GOOS == "windows"
	default:
		return false
	}
}

// Remove deletes a file resource identified by the url.  other resource types return an error.
func Remove(ctx context.Context, url *url.URL) error {
	if !IsWriteSupported(url) {
		return fmt.Errorf("scheme %s is not supported", url.Scheme)
	}

	switch url.Scheme {
	case "file":
		path, err := URLToPath(url)
		if err != nil {
			return err
		}

		// path := url.Path
		// if url.Host != "" {
		// 	path = "//" + url.Host + url.Path
		// }
		// if len(path) > 1 && isWindows(path[1:]) {
		// 	path = path[1:]
		// }
		// path = filepath.FromSlash(path)

		stat, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "stat %s:", path)
		}

		// Already deleted
		if stat == nil {
			return nil
		}

		if stat.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		return err
	default:
		panic("IsWriteSupported bug")
	}
}

// OpenAppend opens a resource for append.
func OpenAppend(ctx context.Context, url *url.URL, perm os.FileMode) (io.WriteCloser, error) {
	if !IsWriteSupported(url) {
		return nil, fmt.Errorf("scheme %s is not supported", url.Scheme)
	}

	switch url.Scheme {
	case "file":
		osPath, err := URLToPath(url)
		if err != nil {
			return nil, err
		}
		return os.OpenFile(osPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm)
	default:
		panic("IsWriteSupported bug")
	}
}

// OpenTruncate opens a a resource and truncates it.
func OpenTruncate(ctx context.Context, url *url.URL, perm os.FileMode) (io.WriteCloser, error) {
	if !IsWriteSupported(url) {
		return nil, fmt.Errorf("scheme %s is not supported", url.Scheme)
	}

	switch url.Scheme {
	case "file":
		osPath, err := URLToPath(url)
		if err != nil {
			return nil, err
		}
		return os.OpenFile(osPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	default:
		panic("IsWriteSupported bug")
	}
}
