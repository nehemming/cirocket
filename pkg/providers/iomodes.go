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

const (
	// IOModeInput file can be used for input.
	IOModeInput = IOMode(1 << iota)

	// IOModeOutput file can be used for output.
	IOModeOutput

	// IOModeError file can be used for errors.
	IOModeError

	// IOModeTruncate file should be truncated.
	IOModeTruncate

	// IOModeAppend file should be appended to.
	IOModeAppend

	// IOModeNone is the default empty mde.
	IOModeNone = IOMode(0)
)

var ioModeMap = map[IOMode]rune{
	IOModeInput:    'i',
	IOModeOutput:   'o',
	IOModeError:    'e',
	IOModeAppend:   'a',
	IOModeTruncate: 't',
}

var ioAllModes = []IOMode{
	IOModeInput,
	IOModeOutput,
	IOModeError,
	IOModeAppend,
	IOModeTruncate,
}

// String converts IOMode to a string representation.
func (mode IOMode) String() string {
	runes := make([]rune, len(ioAllModes))

	for i, v := range ioAllModes {
		if mode&v == v {
			runes[i] = ioModeMap[v]
		} else {
			runes[i] = '-'
		}
	}

	return string(runes)
}
