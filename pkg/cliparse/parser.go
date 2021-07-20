// Package cliparse converts a command line string into CommandLine program and args struct
package cliparse

import (
	"errors"
	"strings"
)

type (
	// Parser is a basic command line parser.
	// Command lines are converted to Commandline instance, containing the program and an array or arguments.
	// Parser handles string expansion with environment variables and respects quotation
	// the parse can also glob wildcards in arguments.
	Parser struct {
		glob func(string) ([]string, error)
	}

	// Commandline represents a program and arguments.
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

// NewParse creates a new parser.
func NewParse() *Parser {
	return &Parser{}
}

// WithGlob adds a glob function too the parser, this expands
// any wildcard arguments.
func (p *Parser) WithGlob(glob func(string) ([]string, error)) *Parser {
	p.glob = glob
	return p
}

// Parse parses a command line or returns an error
// The command line may contain space separated args.  Additional can also
// be provided.
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
	}

	return append(args, arg), nil
}

type parseWorker struct {
	args            []string
	inArgs, inSpace bool
	argStart        int
	mode            quoteMode
	programPath     string
}

func (pw *parseWorker) parseArgsInsideQuotes(i int, r rune, line []rune) error {
	switch r {
	case ' ':
		if pw.inSpace {
			return nil
		}
		pw.inSpace = true
		// found an arg boundary
		pw.args = append(pw.args, string(line[pw.argStart:i]))
	case '|', '>', '<', '&', ';':
		// avoid any shell characters not supported
		return errors.New("unsupported shell expression")
	case '"':
		// Enter quote mode
		pw.mode = doubleQuotes
	case '\'':
		// Enter quote mode
		pw.mode = singleQuotes
	}

	// if character exits space mode
	if pw.inSpace && r != ' ' {
		pw.inSpace = false
		pw.argStart = i
	}

	return nil
}

func (pw *parseWorker) parseProgramInsideQuotes(i int, r rune, line []rune) error {
	// In program mode
	switch r {
	case ' ':
		// Found end of program
		pw.programPath = string(line[:i])
		pw.inSpace, pw.inArgs = true, true
	case '*', '?': // no wildcards in program
		fallthrough
	case '|', '>', '<', '&', ';':
		return errors.New("unsupported shell expression")
	case '"':
		pw.mode = doubleQuotes
	case '\'':
		pw.mode = singleQuotes
	}

	return nil
}

func (pw *parseWorker) parseInsideQuotes(i int, r rune, line []rune) error {
	// check if in arg or program section of line
	if pw.inArgs {
		return pw.parseArgsInsideQuotes(i, r, line)
	}
	return pw.parseProgramInsideQuotes(i, r, line)
}

func (pw *parseWorker) completeParse(line []rune) (string, []string, error) {
	// Check did not exit parse still inside a quote
	if pw.mode == singleQuotes {
		return "", nil, errors.New("closing single quote missing")
	}
	if pw.mode == doubleQuotes {
		return "", nil, errors.New("closing double quote missing")
	}

	// as timed training space, last character must exit a program or an arg
	if pw.inArgs {
		pw.args = append(pw.args, string(line[pw.argStart:]))
	} else {
		pw.programPath = string(line)
	}

	pw.programPath, _ = cleanQuotes(pw.programPath, true)

	return pw.programPath, pw.args, nil
}

func parseLine(commandLine string) (string, []string, error) {
	// convert the line to rune's
	line := []rune(strings.Trim(commandLine, " "))

	pw := &parseWorker{
		args: make([]string, 0),
	}

	// walk the line
	for i, r := range line {
		// if not inside quotes
		if pw.mode == noQuotes {
			if err := pw.parseInsideQuotes(i, r, line); err != nil {
				return "", nil, err
			}
		} else {
			switch r {
			case '"':
				if pw.mode == doubleQuotes {
					pw.mode = noQuotes
				}
			case '\'':
				if pw.mode == singleQuotes {
					pw.mode = noQuotes
				}
			}
		}
	}

	return pw.completeParse(line)
}

type globWorker struct {
	clean    []rune
	globMode int
	mode     quoteMode
}

func (gw *globWorker) buildCleanOutsideQuotes(r rune, stripQuotes bool) {
	switch r {
	case '*', '?':
		if gw.globMode == 0 {
			gw.globMode = 1
		}
		gw.clean = append(gw.clean, r)
	case '"':
		gw.mode = doubleQuotes
		if !stripQuotes {
			gw.clean = append(gw.clean, r)
		}
	case '\'':
		gw.mode = singleQuotes
		if !stripQuotes {
			gw.clean = append(gw.clean, r)
		}
	default:
		gw.clean = append(gw.clean, r)
	}
}

func (gw *globWorker) buildCleanInsideQuotes(r rune, stripQuotes bool) {
	switch r {
	case '*', '?':
		gw.globMode = -1
		gw.clean = append(gw.clean, r)
	case '"':
		if gw.mode == doubleQuotes {
			gw.mode = noQuotes
		}
		if !stripQuotes {
			gw.clean = append(gw.clean, r)
		}
	case '\'':
		if gw.mode == singleQuotes {
			gw.mode = noQuotes
		}
		if !stripQuotes {
			gw.clean = append(gw.clean, r)
		}
	default:
		gw.clean = append(gw.clean, r)
	}
}

func cleanQuotes(arg string, stripQuotes bool) (string, int) {
	line := []rune(arg)

	gw := &globWorker{
		clean: make([]rune, 0, len(line)),
	}

	for _, r := range line {
		if gw.mode == noQuotes {
			gw.buildCleanOutsideQuotes(r, stripQuotes)
		} else {
			gw.buildCleanInsideQuotes(r, stripQuotes)
		}
	}

	// Line parsed in clean runes
	return string(gw.clean), gw.globMode
}
