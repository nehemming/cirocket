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
	"context"
	"io"
	"testing"
	"time"
)

func TestNewURLProviderBlankErrors(t *testing.T) {
	p, err := NewURLProvider("", time.Second*5, false)
	if err == nil || p != nil {
		t.Error("error expected", err, p)
	}
}

func TestNewURLProvider(t *testing.T) {
	p, err := NewURLProvider("somewhere", 0, false)
	if err != nil || p == nil {
		t.Error("error unexpected", err, p)
	}

	up := p.(*urlResourceProvider)
	if up.url != "somewhere" || up.timeout != time.Second*30 {
		t.Error("values unexpected timeout default", up.url, up.timeout)
	}

	p, err = NewURLProvider("somewhere", time.Second*5, false)
	if err != nil || p == nil {
		t.Error("error unexpected", err, p)
	}

	up = p.(*urlResourceProvider)

	if up.url != "somewhere" || up.timeout != time.Second*5 {
		t.Error("values unexpected", up.url, up.timeout)
	}

	// Check write denied too
	_, err = p.OpenWrite(context.Background())
	if err == nil {
		t.Error("error expected for write open ", err)
	}
}

func TestNReadURLProvider(t *testing.T) {
	p, err := NewURLProvider("https://raw.githubusercontent.com/nehemming/cirocket/master/README.md", time.Second*10, false)
	if err != nil || p == nil {
		t.Error("error unexpected", err, p)
	}
	ctx := context.Background()

	w, err := p.OpenRead(ctx)
	if err != nil {
		t.Error("open issue", err)
		return
	}
	b, err := io.ReadAll(w)
	if err != nil {
		t.Error("read issue", err)
		return
	}
	defer w.Close()

	if len(b) < 100 {
		t.Error("reading b", len(b))
	}
}

func TestNReadURLProviderOptional(t *testing.T) {
	p, err := NewURLProvider("https://raw.githubusercontent.com/nehemming/cirocket/master/README-ntfound.md", time.Second*10, true)
	if err != nil || p == nil {
		t.Error("error unexpected", err, p)
	}
	ctx := context.Background()

	w, err := p.OpenRead(ctx)
	if err != nil {
		t.Error("open issue", err)
		return
	}
	b, err := io.ReadAll(w)
	if err != nil {
		t.Error("read issue", err)
		return
	}
	defer w.Close()

	if len(b) != 0 {
		t.Error("reading b", len(b))
	}
}

func TestNReadURLProviderErrorsNotFound(t *testing.T) {
	p, err := NewURLProvider("https://raw.githubusercontent.com/nehemming/cirocket/master/README-ntfound.md", time.Second*10, false)
	if err != nil || p == nil {
		t.Error("error unexpected", err, p)
	}
	ctx := context.Background()

	_, err = p.OpenRead(ctx)
	if err == nil {
		t.Error("no open issue")
	}
}
