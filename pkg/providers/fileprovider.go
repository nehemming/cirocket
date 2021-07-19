package providers

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
)

type (
	FileDetail interface {
		FilePath() string
		IOMode() IOMode
		FileMode() os.FileMode
		InMode(mode IOMode) bool
	}

	fileResourceProvider struct {
		filePath string
		ioMode   IOMode
		fileMode os.FileMode
		optional bool
	}

	nopReaderCloser struct {
		reader io.Reader
	}

	nopWriterCloser struct {
		writer io.Writer
	}
)

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

	return &fileResourceProvider{
		filePath: path,
		ioMode:   ioMode,
		fileMode: fileMode,
		optional: optional,
	}, nil
}

func (rc *nopReaderCloser) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return nil, errors.New("output is not supported")
}

func (rc *nopReaderCloser) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return rc, nil
}

func (rc *nopReaderCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *nopReaderCloser) Close() error {
	return nil
}

func (wc *nopWriterCloser) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return wc, nil
}

func (wc *nopWriterCloser) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return nil, errors.New("input is not supported")
}

func (wc *nopWriterCloser) Write(p []byte) (n int, err error) {
	return wc.writer.Write(p)
}

func (wc *nopWriterCloser) Close() error {
	return nil
}

func (fp *fileResourceProvider) FilePath() string {
	return fp.filePath
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

	rc, err := os.Open(fp.filePath)
	if err != nil {
		if !os.IsNotExist(err) || !fp.optional {
			return nil, err
		}

		// Return an empty reader
		return &nopReaderCloser{
			reader: bytes.NewBufferString(""),
		}, nil
	}

	return rc, err
}

func (fp *fileResourceProvider) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	if fp.ioMode&(IOModeOutput|IOModeError) == IOModeNone {
		return nil, errors.New("output is not supported")
	}

	if (fp.ioMode & IOModeTruncate) == IOModeTruncate {
		return os.OpenFile(fp.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fp.fileMode)
	} else if (fp.ioMode & IOModeAppend) == IOModeAppend {
		return os.OpenFile(fp.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, fp.fileMode)
	} else {
		panic("validation should have caught missing append or truncate")
	}
}
