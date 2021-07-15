package cliparse

import (
	"path/filepath"
	"testing"
)

func TestParseLineEmptyString(t *testing.T) {

	p, a, err := parseLine("")

	if err != nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnly(t *testing.T) {

	expectedProgram := "run.exe"
	p, a, err := parseLine(expectedProgram)

	if err != nil || p != expectedProgram || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlyRemovesPadding(t *testing.T) {

	expectedProgram := "run.exe"
	p, a, err := parseLine("  " + expectedProgram + "  ")

	if err != nil || p != expectedProgram || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramHasSpacesInQuotes(t *testing.T) {

	expectedProgram := "run my.exe"
	p, a, err := parseLine("  'run my'.exe  ")

	if err != nil || p != expectedProgram || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlySingleQuoted(t *testing.T) {

	expectedProgram := "run.exe"
	p, a, err := parseLine("'run.exe'")

	if err != nil || p != expectedProgram || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlyDoubleQuoted(t *testing.T) {

	expectedProgram := "run.exe"
	p, a, err := parseLine("\"run.exe\"")

	if err != nil || p != expectedProgram || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlyDSingleuotedMissingEnd(t *testing.T) {

	expectedProgram := "'run.exe"
	p, a, err := parseLine(expectedProgram)

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlyDoubleQuotedMissingEnd(t *testing.T) {

	expectedProgram := "\"run.exe"
	p, a, err := parseLine(expectedProgram)

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlySingleQuotedMissingStart(t *testing.T) {

	expectedProgram := "run.exe\""
	p, a, err := parseLine(expectedProgram)

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOnlyDoubleQuotedMissingStart(t *testing.T) {

	expectedProgram := "run.exe\""
	p, a, err := parseLine(expectedProgram)

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramOneArg(t *testing.T) {

	expectedProgram := "run.exe"
	expectedArgs := []string{"one"}
	p, a, err := parseLine(expectedProgram + " one")

	if err != nil || p != expectedProgram || len(a) != len(expectedArgs) {
		t.Error("unexpected", p, a, err)
		return
	}

	for i, v := range a {
		if v != expectedArgs[i] {
			t.Error("arg issue", i, v, expectedArgs[i])
		}
	}
}

func TestParseLineProgramThreeArg(t *testing.T) {

	expectedProgram := "run.exe"
	expectedArgs := []string{"one", "two", "three"}
	p, a, err := parseLine(expectedProgram + " one two three")

	if err != nil || p != expectedProgram || len(a) != len(expectedArgs) {
		t.Error("unexpected", p, a, err)
		return
	}

	for i, v := range a {
		if v != expectedArgs[i] {
			t.Error("arg issue", i, v, expectedArgs[i])
		}
	}
}

func TestParseLineProgramThreeArgQuoted(t *testing.T) {

	expectedProgram := "run.exe"
	expectedArgs := []string{"one", "\"two\"", "'three'"}
	p, a, err := parseLine(expectedProgram + " one \"two\" 'three'")

	if err != nil || p != expectedProgram || len(a) != len(expectedArgs) {
		t.Error("unexpected", p, a, err)
		return
	}

	for i, v := range a {
		if v != expectedArgs[i] {
			t.Error("arg issue", i, v, expectedArgs[i])
		}
	}
}

func TestParseLineProgramTwoArgQuoted(t *testing.T) {

	expectedProgram := "run.exe"
	expectedArgs := []string{"one", "\"two 'three'\""}
	p, a, err := parseLine(expectedProgram + " one \"two 'three'\"")

	if err != nil || p != expectedProgram || len(a) != len(expectedArgs) {
		t.Error("unexpected", p, a, err)
		return
	}

	for i, v := range a {
		if v != expectedArgs[i] {
			t.Error("arg issue", i, v, expectedArgs[i])
		}
	}
}

func TestParseLineProgramTwoArgQuoted2(t *testing.T) {

	expectedProgram := "run.exe"
	expectedArgs := []string{"one", "'two \"thr\"ee'"}
	p, a, err := parseLine(expectedProgram + "  one 'two \"thr\"ee'")

	if err != nil || p != expectedProgram || len(a) != len(expectedArgs) {
		t.Error("unexpected", p, a, err)
		return
	}

	for i, v := range a {
		if v != expectedArgs[i] {
			t.Error("arg issue", i, v, expectedArgs[i])
		}
	}
}

func TestParseLineProgramOnlyWithIllegalShellChar(t *testing.T) {

	p, a, err := parseLine("run*.exe|")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}

	p, a, err = parseLine("run.exe|")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}

	p, a, err = parseLine("run.exe<")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}

	p, a, err = parseLine("run.exe>>")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}

}

func TestParseLineProgramWitArgsWithPipe(t *testing.T) {

	p, a, err := parseLine("run.exe | two")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestParseLineProgramWitArgsWithLeadingChar(t *testing.T) {

	p, a, err := parseLine(">run.exe|two")

	if err == nil || p != "" || len(a) != 0 {
		t.Error("unexpected", p, a, err)
	}
}

func TestCleanQuotesBlank(t *testing.T) {

	clean, globMode := cleanQuotes("", true)

	if clean != "" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("", false)

	if clean != "" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestCleanQuotesNoQuotes(t *testing.T) {

	clean, globMode := cleanQuotes("hello", true)

	if clean != "hello" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("hello", false)

	if clean != "hello" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestCleanQuotesNoWithQuotes(t *testing.T) {

	clean, globMode := cleanQuotes("'hello'", true)

	if clean != "hello" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("'hello'", false)

	if clean != "'hello'" && globMode != 0 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestCleanQuotesWildcardNoQuotes(t *testing.T) {

	clean, globMode := cleanQuotes("hello*me", true)

	if clean != "hello*me" && globMode != 1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("hello*me", false)

	if clean != "hello*me" && globMode != -1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestCleanQuotesWildcardQuotes(t *testing.T) {

	clean, globMode := cleanQuotes("\"hello\"*me", true)

	if clean != "hello*me" && globMode != 1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("\"hello\"*me", false)

	if clean != "\"hello\"*me" && globMode != -1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestCleanQuotesWildcardIndideQuotes(t *testing.T) {

	clean, globMode := cleanQuotes("\"hello*\"me", true)

	if clean != "hello*me" && globMode != -1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}

	clean, globMode = cleanQuotes("\"hello*\"me", false)

	if clean != "\"hello*\"me" && globMode != -1 {
		t.Error("cleanQuotes blank in unexpected result", clean, globMode)
	}
}

func TestParserCreate(t *testing.T) {

	p := NewParse()

	if p.glob != nil {
		t.Error("Glob not nil on by default")
	}

	x := p.WithGlob(nil)

	if p.glob != nil {
		t.Error("Glob not nil on by default")
	}

	if x != p {
		t.Error("WithGlob fluency issue")
	}

	p.WithGlob(filepath.Glob)

	if p.glob == nil {
		t.Error("Glob nil after WithGlob")
	}
}

func TestExpandArgIgnoresEmpty(t *testing.T) {

	p := NewParse()

	slice := make([]string, 0)

	slice, err := p.expandArg(slice, "", true)

	if err != nil || len(slice) > 0 {
		t.Error("expandArg return unexpected", len(slice), err)
	}
}

func TestExpandArgReturnsArg(t *testing.T) {

	p := NewParse()

	slice := make([]string, 0)

	slice, err := p.expandArg(slice, "hello", true)

	if err != nil || len(slice) != 1 || slice[0] != "hello" {
		t.Error("expandArg return unexpected", len(slice), err)
	}
}

func TestExpandArgGlobs(t *testing.T) {

	p := NewParse().WithGlob(filepath.Glob)

	slice := make([]string, 0)

	slice, err := p.expandArg(slice, "*", true)

	if err != nil || len(slice) != 2 {
		t.Errorf("expandArg return unexpected %v %v", slice, err)
	}
}

func TestCombineArgsNoAdditionSucceeds(t *testing.T) {

	p := NewParse().WithGlob(filepath.Glob)

	args := []string{"'hello'", "'there'"}

	slice, err := p.combineArgs(args)
	if err != nil || len(slice) != 2 && slice[1] == "there" {
		t.Errorf("combineArgs return unexpected %v %v", slice, err)
	}
}

func TestCombineArgsAdditionSucceeds(t *testing.T) {

	p := NewParse().WithGlob(filepath.Glob)

	args := []string{"'hello'", "'there'"}

	slice, err := p.combineArgs(args, "this", "'is quoted'")
	if err != nil || len(slice) != 4 && slice[3] == "'is quoted'" {
		t.Errorf("combineArgs return unexpected %v %v", slice, err)
	}
}

func TestCombineArgsAdditionNoGlobSucceeds(t *testing.T) {

	p := NewParse().WithGlob(filepath.Glob)

	args := []string{"'*'", "'there'"}

	slice, err := p.combineArgs(args, "this", "'*'")
	if err != nil || len(slice) != 4 && slice[0] == "*" && slice[3] == "'*'" {
		t.Errorf("combineArgs return unexpected %v %v", slice, err)
	}
}

func TestCombineArgsAdditionGlobsSucceeds(t *testing.T) {

	p := NewParse().WithGlob(filepath.Glob)

	args := []string{"*", "'there'"}

	slice, err := p.combineArgs(args, "this", "*")
	if err != nil || len(slice) != 8 && slice[0] != "*" && slice[3] == "there" {
		t.Errorf("combineArgs return unexpected %v %v", slice, err)
	}
}

func TestParseSucceeds(t *testing.T) {

	c, err := NewParse().WithGlob(filepath.Glob).Parse("")

	if err != nil || len(c.Args) > 0 || c.ProgramPath != "" {
		t.Errorf("Parse return unexpected %v %v", c, err)
	}
}

func TestParseNoGlobSucceeds(t *testing.T) {

	c, err := NewParse().WithGlob(filepath.Glob).Parse("hello there \"this spans\" '*'")

	if err != nil || len(c.Args) != 3 || c.ProgramPath != "hello" {
		t.Errorf("Parse return unexpected %v %v", c, err)
	}
}

func TestParseGlobsSucceeds(t *testing.T) {

	c, err := NewParse().WithGlob(filepath.Glob).Parse("hello there \"this spans\" *")

	if err != nil || len(c.Args) != 4 || c.ProgramPath != "hello" {
		t.Errorf("Parse return unexpected %v %v", c, err)
	}
}
