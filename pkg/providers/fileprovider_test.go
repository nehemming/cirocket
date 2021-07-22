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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nehemming/cirocket/pkg/resource"
)

func TestNewNonClosingReaderProvider(t *testing.T) {
	buf := bytes.NewBufferString("hello")

	provider := NewNonClosingReaderProvider(buf)
	if provider == nil {
		t.Error("unexpected nil provider")
	}

	w, err := provider.OpenWrite(context.Background())
	if err == nil || w != nil {
		t.Error("error expected writer", err, w)
	}

	r, err := provider.OpenRead(context.Background())
	if err != nil || r == nil {
		t.Error("error unexpected reader", err, r)
	}
	r.Close()

	b, err := io.ReadAll(r)
	if err != nil || string(b) != "hello" {
		t.Error("error unexpected read", err, string(b))
	}
}

func TestNewNonClosingWriterProvider(t *testing.T) {
	buf := new(bytes.Buffer)

	provider := NewNonClosingWriterProvider(buf)
	if provider == nil {
		t.Error("unexpected nil provider")
	}

	r, err := provider.OpenRead(context.Background())
	if err == nil || r != nil {
		t.Error("error expected reader", err, r)
	}

	w, err := provider.OpenWrite(context.Background())
	if err != nil || w == nil {
		t.Error("error unexpected writer", err, r)
	}

	w.Close()

	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Error("error unexpected write", err, n)
	}

	s := buf.String()
	if s != "hello" {
		t.Error("error unexpected read", err, s)
	}
}

func TestNewNewFileProviderWrongType(t *testing.T) {
	_, err := NewFileProvider("fileprovider.go", IOModeOutput, 0, false)
	if err == nil {
		t.Error("expected error", err)
	}
	if err.Error() != "neither truncate nor append have been specified, please select only one" {
		t.Error("error mismatch", err)
	}
}

func TestNewNewFileProviderWrongTypeError(t *testing.T) {
	_, err := NewFileProvider("fileprovider.go", IOModeError, 0, false)
	if err == nil {
		t.Error("expected error", err)
	}
	if err.Error() != "neither truncate nor append have been specified, please select only one" {
		t.Error("error mismatch", err)
	}
}

func TestNewNewFileProviderModesError(t *testing.T) {
	_, err := NewFileProvider("fileprovider.go", IOModeError|IOModeAppend|IOModeTruncate, 0, false)
	if err == nil {
		t.Error("expected error", err)
	}
	if err.Error() != "both truncate and append have been specified, please select only one" {
		t.Error("error mismatch", err)
	}
}

func TestNewNewFileProviderNoIOModeError(t *testing.T) {
	_, err := NewFileProvider("fileprovider.go", IOModeAppend|IOModeTruncate, 0, false)
	if err == nil {
		t.Error("expected error", err)
	}
	if err.Error() != "mode is neither input nor output" {
		t.Error("error mismatch", err)
	}
}

func TestNewNewFileProviderBlankPath(t *testing.T) {
	_, err := NewFileProvider("", IOModeAppend|IOModeTruncate, 0, false)
	if err == nil {
		t.Error("expected error", err)
	}
	if err.Error() != "path is blank" {
		t.Error("error mismatch", err)
	}
}

func TestNewFileProviderOpensForRead(t *testing.T) {
	rp, err := NewFileProvider("fileprovider.go", IOModeInput, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	reader, err := rp.OpenRead(context.Background())
	if err != nil {
		t.Error("unexpected error", err)
	}

	defer reader.Close()

	b, err := io.ReadAll(reader)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if !strings.HasPrefix(string(b), "/*") {
		t.Error("unexpected error", err, string(b))
	}
}

func TestNewFileProviderOpensForReaderrorNotFound(t *testing.T) {
	rp, err := NewFileProvider("no_fileprovider.go", IOModeInput, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	_, err = rp.OpenRead(context.Background())
	if err == nil {
		t.Error("expected error")
	}
}

func TestNewFileProviderOpensForReadOptional(t *testing.T) {
	rp, err := NewFileProvider("no_fileprovider.go", IOModeInput, 0, true)
	if err != nil {
		t.Error("unexpected error", err)
	}

	reader, err := rp.OpenRead(context.Background())
	if err != nil {
		t.Error("unexpected error", err)
		return
	}

	defer reader.Close()

	b, err := io.ReadAll(reader)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(b) != 0 {
		t.Error("non zero missing optional ", len(b))
	}
}

func TestNewNewFileProviderErrorsForWrite(t *testing.T) {
	rp, err := NewFileProvider("fileprovider.go", IOModeInput, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	_, err = rp.OpenWrite(context.Background())
	if err == nil {
		t.Error("expected error", err)
	}

	if err.Error() != "output is not supported" {
		t.Error("expected error", err)
	}
}

func TestNewNewFileProviderErrorsForRead(t *testing.T) {
	err := os.MkdirAll("testdata", 0777)
	if err != nil {
		panic(err)
	}

	rp, err := NewFileProvider("testdata/new.dat", IOModeError|IOModeTruncate, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	_, err = rp.OpenRead(context.Background())
	if err == nil {
		t.Error("expected error", err)
	}

	if err.Error() != "input is not supported" {
		t.Error("expected error", err)
	}
}

func TestNewNewFileProviderTruncatesFile(t *testing.T) {
	err := os.MkdirAll("testdata", 0777)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.FromSlash("testdata/trunc.dat"), []byte("bad"), 0666)
	if err != nil {
		panic(err)
	}

	rp, err := NewFileProvider(filepath.FromSlash("testdata/trunc.dat"), IOModeOutput|IOModeTruncate, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	w, err := rp.OpenWrite(context.Background())
	if err != nil {
		t.Error("expected error open", err)
	}

	_, err = w.Write([]byte("good"))
	if err != nil {
		t.Error("expected error write", err)
	}

	b, err := os.ReadFile(filepath.FromSlash("testdata/trunc.dat"))
	if err != nil {
		panic(err)
	}

	if string(b) != "good" {
		t.Error("expected write data", string(b))
	}
}

func TestNewNewFileProviderAppendFile(t *testing.T) {
	err := os.MkdirAll("testdata", 0777)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.FromSlash("testdata/append.dat"), []byte("hello "), 0666)
	if err != nil {
		panic(err)
	}

	rp, err := NewFileProvider(filepath.FromSlash("testdata/append.dat"), IOModeOutput|IOModeAppend, 0, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	w, err := rp.OpenWrite(context.Background())
	if err != nil {
		t.Error("expected error open", err)
	}

	_, err = w.Write([]byte("good"))
	if err != nil {
		t.Error("expected error write", err)
	}

	b, err := os.ReadFile(filepath.FromSlash("testdata/append.dat"))
	if err != nil {
		panic(err)
	}

	if string(b) != "hello good" {
		t.Error("expected write data", string(b))
	}

	err = os.WriteFile(filepath.FromSlash("testdata/append.dat"), []byte("hello "), 0666)
	if err != nil {
		panic(err)
	}
}

func TestNewNewFileProviderFileDetails(t *testing.T) {
	res, err := NewFileProvider("fileprovider.go", IOModeInput, 0666, false)
	if err != nil {
		t.Error("unexpected error", err)
	}

	fp, ok := res.(FileDetailer)
	if !ok {
		t.Error("error FileDetailer not implemented")
		return
	}

	wd, _ := os.Getwd()
	path := strings.TrimPrefix(fp.FilePath(), wd)
	if path != "/fileprovider.go" {
		t.Error("file path", fp.FilePath(), path)
	}

	// using url is easier
	path, _ = resource.URLToRelativePath(fp.URL())
	if path != "fileprovider.go" {
		t.Error("file path", fp.URL(), path)
	}

	if fp.IOMode() != IOModeInput {
		t.Error("io mode", fp.IOMode())
	}

	if fp.FileMode() != 0666 {
		t.Error("mode", fp.FileMode())
	}
}
