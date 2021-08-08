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
	"net/url"
	"runtime"
	"testing"
)

func TestIsWriteSupportedWeb(t *testing.T) {
	if IsWriteSupported(new(url.URL)) {
		t.Error("empty can write")
	}

	u, e := url.Parse("https://news.site/daily")
	if e != nil {
		panic(e)
	}

	if IsWriteSupported(u) {
		t.Error("web can write")
	}
}

func TestIsWriteSupportedFile(t *testing.T) {
	u, e := url.Parse("file://news.site/daily")
	if e != nil {
		panic(e)
	}
	if IsWriteSupported(u) && runtime.GOOS != "windows" {
		t.Error("file with server write")
	}

	u, e = url.Parse("file:///news.site/daily")
	if e != nil {
		panic(e)
	}
	if !IsWriteSupported(u) {
		t.Error("local cannot write")
	}

	u, e = url.Parse("file:/news.site/daily")
	if e != nil {
		panic(e)
	}
	if !IsWriteSupported(u) {
		t.Error("local simple cannot write")
	}
}

func TestWriterFileTruncate(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("testdata", "create.tmp")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	data := "hello there"

	wc, err := OpenTruncate(ctx, u, 0666)
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	// try write
	_, err = wc.Write([]byte(data))
	if err != nil {
		t.Error("unexpected", err)
		wc.Close()
		return
	}

	wc.Close()

	b, err := ReadURL(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	} else if string(b) != data {
		t.Error("unexpected", string(b))
	}

	// Clean up
	err = Remove(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestWriterFileAppend(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("testdata", "append.tmp")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	data := "hello"

	err = Remove(ctx, u)
	if err != nil {
		t.Error("unexpected", err, u)
		return
	}

	for i := 0; i < 2; i++ {
		wc, err := OpenAppend(ctx, u, 0666)
		if err != nil {
			t.Error("unexpected", err)
			return
		}

		// try write
		_, err = wc.Write([]byte(data))
		if err != nil {
			t.Error("unexpected", err)
			wc.Close()
			return
		}

		wc.Close()
	}

	b, err := ReadURL(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	} else if string(b) != data+data {
		t.Error("unexpected", string(b))
	}

	// Clean up
	err = Remove(ctx, u)
	if err != nil {
		t.Error("unexpected", err)
	}
}

func TestWriterFileAppendSchemeFail(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("https://testdata", "append.tmp")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	_, err = OpenAppend(ctx, u, 0666)
	if err == nil {
		t.Error("expected an error")
		return
	}
}

func TestWriterFileTruncateSchemeFail(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("https://testdata", "create.tmp")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	_, err = OpenTruncate(ctx, u, 0666)
	if err == nil {
		t.Error("expected an error")
		return
	}
}

func TestRemoveSchemeFail(t *testing.T) {
	ctx := context.Background()

	u, err := UltimateURL("https://testdata", "create.tmp")
	if err != nil {
		t.Error("unexpected", err)
		return
	}

	err = Remove(ctx, u)
	if err == nil {
		t.Error("expected an error")
		return
	}
}
