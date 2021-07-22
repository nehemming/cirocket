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
	"bufio"
	"context"
	"io"

	"github.com/pkg/errors"
)

type (
	// LineFunc is a callback function type that is called per line  written to a line provider.
	LineFunc func(string)

	streamer struct {
		lineFunc LineFunc
		writer   *io.PipeWriter
		done     chan struct{}
	}
)

// NewLineProvider writes lines of text to a output function.
func NewLineProvider(fn func(string)) ResourceProvider {
	return LineFunc(fn)
}

// OpenRead returns an error if called as reads are not possible from a line provider.
func (lf LineFunc) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return nil, errors.New("input is not supported")
}

// OpenWrite opens resource for writing.
func (lf LineFunc) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	reader, writer := io.Pipe()

	st := &streamer{
		lineFunc: lf,
		writer:   writer,
		done:     make(chan struct{}),
	}

	// Start reader, runs until pipe is closed
	go func() {
		defer close(st.done)

		in := bufio.NewScanner(reader)

		for in.Scan() {
			st.lineFunc(in.Text())
		}
	}()

	return st, nil
}

func (st *streamer) Write(p []byte) (n int, err error) {
	return st.writer.Write(p)
}

func (st *streamer) Close() error {
	st.writer.Close()

	// Wait for close to complete
	<-st.done
	return nil
}
