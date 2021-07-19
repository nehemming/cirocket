package providers

import "testing"

func TestIdempotentCloserTest(t *testing.T) {
	tc := &testCloser{}

	idc := NewIdempotentCloser(tc)

	idc.Close()

	if tc.counter != 1 {
		t.Error("not closed")
	}

	idc.Close()

	if tc.counter != 1 {
		t.Error("over closed")
	}
}
