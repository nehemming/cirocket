package providers

import (
	"context"
	"testing"
)

func TestNewLineProvider(t *testing.T) {
	called := new(bool)
	*called = false

	fn := func(s string) {
		*called = true
		if s != "hello" {
			t.Error("expected hello", s)
		}
	}

	provider := NewLineProvider(fn)

	// Check read fails
	_, err := provider.OpenRead(context.Background())
	if err == nil {
		t.Error("error expected")
	}

	writer, err := provider.OpenWrite(context.Background())
	if err != nil {
		t.Error("error unexpected", err)
		return
	}

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Error("error unexpected", err)
		writer.Close()
		return
	}

	writer.Close()

	if !*called {
		t.Error("not called fn")
	}
}
