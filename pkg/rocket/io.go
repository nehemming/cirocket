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
