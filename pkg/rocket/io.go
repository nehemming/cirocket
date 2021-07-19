package rocket

import (
	"github.com/nehemming/cirocket/pkg/providers"
)

const (
	// InputIO is the input file key.
	InputIO = providers.ResourceID("input")
	// OutputIO is the output file key.
	OutputIO = providers.ResourceID("output")
	// ErrorIO is the error file key.
	ErrorIO = providers.ResourceID("error")

	// Stdin is the Std in resource.
	Stdin = providers.ResourceID("stdin")
	// Stdout is the Std out resource.
	Stdout = providers.ResourceID("stdout")
	// Stderr is the Std error resource.
	Stderr = providers.ResourceID("stderr")
)
