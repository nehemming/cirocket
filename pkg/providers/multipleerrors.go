package providers

import (
	"fmt"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
)

func multiErrorFormatter(es []error) string {
	if len(es) == 1 {
		return es[0].Error()
	}

	text := make([]string, len(es))
	for i, err := range es {
		text[i] = fmt.Sprintf("%s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred: %s",
		len(es), strings.Join(text, "; "))
}

func bindMultiErrorFormatting(err error) error {
	if err != nil {
		if multi, ok := err.(*multierror.Error); ok {
			multi.ErrorFormat = multiErrorFormatter
		}
	}

	return err
}
