package rocket

import (
	"context"
	"testing"
)

func TestContextRoundTrip(t *testing.T) {
	capComm := newCapCommFromEnvironment("dir/file")

	ctx := NewContextWithCapComm(context.Background(), capComm)

	ret := GetCapCommContext(ctx)

	if ret != capComm {
		t.Error("No round trip")
	}
}

func TestGetCapCommContextWithNoneSet(t *testing.T) {
	ret := GetCapCommContext(context.Background())

	if ret != nil {
		t.Error("ret should be nil")
	}
}
