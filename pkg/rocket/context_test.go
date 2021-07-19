package rocket

import (
	"context"
	"testing"

	"github.com/nehemming/cirocket/pkg/loggee/stdlog"
)

func TestContextRoundTrip(t *testing.T) {
	capComm := newCapCommFromEnvironment("dir/file", stdlog.New())

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
