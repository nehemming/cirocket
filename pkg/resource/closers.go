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
	"io"
)

type (
	nopReaderCloser struct {
		reader io.Reader
	}

	nopWriterCloser struct {
		writer io.Writer
	}
)

// NewReadCloser creates a reader closer from a reader.  The close operation is a no op.
// This type can be used when a interface requires a reader closer.
func NewReadCloser(reader io.Reader) io.ReadCloser {
	return &nopReaderCloser{reader}
}

func (rc *nopReaderCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *nopReaderCloser) Close() error {
	return nil
}

// NewWriteCloser creates a writer closer from a reader.  The close operation is a no op.
// This type can be used when a interface requires a reader closer.
func NewWriteCloser(writer io.Writer) io.WriteCloser {
	return &nopWriterCloser{writer}
}

func (wc *nopWriterCloser) Write(p []byte) (n int, err error) {
	return wc.writer.Write(p)
}

func (wc *nopWriterCloser) Close() error {
	return nil
}
