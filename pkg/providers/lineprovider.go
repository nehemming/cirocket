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
