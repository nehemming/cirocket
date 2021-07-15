package cliparse

import (
	"errors"
	"strings"
)

type (
	Parser struct {
		glob func(string) ([]string, error)
	}

	Commandline struct {
		ProgramPath string
		Args        []string
	}

	quoteMode int
)

const (
	noQuotes = quoteMode(iota)
	singleQuotes
	doubleQuotes
)

// NewParse creates a new parser
func NewParse() *Parser {
	return &Parser{}
}

// adds a glob function too the parser, this expands
// any wildcard arguments
func (p *Parser) WithGlob(glob func(string) ([]string, error)) *Parser {
	p.glob = glob
	return p
}

// Parse parses a command line or returns an error
// The command line may contain space seperated args.  Additional can also
// be provided
func (p *Parser) Parse(commandLine string, args ...string) (c *Commandline, err error) {

	// Parse the command line into program path and args
	// args after the call will still coontain any surrounding quotes
	programPath, commandLineArgs, err := parseLine(commandLine)
	if err != nil {
		return
	}

	// Combine command line and additional args together
	// Args will be globed and expanded as necessary too
	args, err = p.combineArgs(commandLineArgs, args...)
	if err != nil {
		return
	}

	c = &Commandline{
		ProgramPath: programPath,
		Args:        args,
	}

	return
}

func (p *Parser) combineArgs(args []string, additional ...string) ([]string, error) {

	combined := make([]string, 0)
	var err error

	// Args
	for _, arg := range args {
		combined, err = p.expandArg(combined, arg, true)
		if err != nil {
			return nil, err
		}
	}

	for _, arg := range additional {
		combined, err = p.expandArg(combined, arg, false)
		if err != nil {
			return nil, err
		}
	}

	return combined, nil
}

func (p *Parser) expandArg(args []string, arg string, stripQuotes bool) ([]string, error) {

	// Ignore empty args
	if arg == "" {
		return args, nil
	}

	// expand arg
	arg, globMode := cleanQuotes(arg, stripQuotes)

	if globMode > 0 && p.glob != nil {
		globArgs, err := p.glob(arg)
		if err != nil {
			return nil, err
		}
		return append(args, globArgs...), nil
	} else {
		return append(args, arg), nil
	}
}

func parseLine(commandLine string) (string, []string, error) {

	// convert the line to rune's
	line := []rune(strings.Trim(commandLine, " "))

	var inArgs, inSpace bool
	var argStart int
	var mode quoteMode
	var programPath string
	args := make([]string, 0)

	// walk the line
	for i, r := range line {

		// if not inside quotes
		if mode == noQuotes {

			// check if in arg or program section of line
			if inArgs {
				switch r {
				case ' ':
					if inSpace {
						continue
					}
					inSpace = true
					// found an arg boundary
					args = append(args, string(line[argStart:i]))
				case '|', '>', '<', '&', ';':
					// avoid any shell characters not supported
					return "", nil, errors.New("unsupported shell expression")
				case '"':
					// Enter quote mode
					mode = doubleQuotes
				case '\'':
					// Enter quote mode
					mode = singleQuotes
				}

				// if character exits space mode
				if inSpace && r != ' ' {
					inSpace = false
					argStart = i
				}
			} else {
				// In program mode
				switch r {
				case ' ':
					// Found end of program
					programPath = string(line[:i])
					inSpace, inArgs = true, true
				case '*', '?': // no wildcards in program
					fallthrough
				case '|', '>', '<', '&', ';':
					return "", nil, errors.New("unsupported shell expression")
				case '"':
					mode = doubleQuotes
				case '\'':
					mode = singleQuotes
				}
			}
		} else {
			switch r {
			case '"':
				if mode == doubleQuotes {
					mode = noQuotes
				}
			case '\'':
				if mode == singleQuotes {
					mode = noQuotes
				}
			}
		}
	}

	// Check did not exit parse still inside a quote
	if mode == singleQuotes {
		return "", nil, errors.New("closing single quote missing")
	}
	if mode == doubleQuotes {
		return "", nil, errors.New("closing double quote missing")
	}

	// as timed training space, last character must exit a program or an arg
	if inArgs {
		args = append(args, string(line[argStart:]))
	} else {
		programPath = string(line)
	}

	programPath, _ = cleanQuotes(programPath, true)

	return programPath, args, nil
}

func cleanQuotes(arg string, stripQuotes bool) (string, int) {
	line := []rune(arg)
	clean := make([]rune, 0, len(line))
	var globMode int
	var mode quoteMode

	for _, r := range line {

		if mode == noQuotes {
			switch r {
			case '*', '?':
				if globMode == 0 {
					globMode = 1
				}
				clean = append(clean, r)
			case '"':
				mode = doubleQuotes
				if !stripQuotes {
					clean = append(clean, r)
				}
			case '\'':
				mode = singleQuotes
				if !stripQuotes {
					clean = append(clean, r)
				}
			default:
				clean = append(clean, r)
			}
		} else {
			switch r {
			case '*', '?':
				globMode = -1
				clean = append(clean, r)
			case '"':
				if mode == doubleQuotes {
					mode = noQuotes
				}
				if !stripQuotes {
					clean = append(clean, r)
				}
			case '\'':
				if mode == singleQuotes {
					mode = noQuotes
				}
				if !stripQuotes {
					clean = append(clean, r)
				}
			default:
				clean = append(clean, r)
			}
		}
	}

	// Line parsed in clean runes
	return string(clean), globMode
}
