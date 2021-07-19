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
