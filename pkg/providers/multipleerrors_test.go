package providers

import (
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func TestMultipleErrorFormattingOnNil(t *testing.T) {
	err := bindMultiErrorFormatting(nil)
	if err != nil {
		t.Error("Mismatch nil")
	}
}

func TestMultipleErrorFormattingSimple(t *testing.T) {
	err := bindMultiErrorFormatting(errors.New("simple"))

	if err.Error() != "simple" {
		t.Error("Mismatch simple", err)
	}

	// err = multierror.Append(err, errors.Wrapf(e, "close[%d]", i))
}

func TestMultipleErrorFormattingSingle(t *testing.T) {
	var err error
	err = multierror.Append(err, errors.New("one"))

	err = bindMultiErrorFormatting(err)

	if err.Error() != "one" {
		t.Error("Mismatch single", err)
	}
}

func TestMultipleErrorFormattingMultiple(t *testing.T) {
	var err error
	err = multierror.Append(err, errors.New("one"))
	err = multierror.Append(err, errors.New("two"))

	err = bindMultiErrorFormatting(err)

	if err.Error() != "2 errors occurred: one; two" {
		t.Error("Mismatch multiple", err)
	}
}
