package rocket

import (
	"fmt"
	"testing"
)

func TestIOModes(t *testing.T) {

	r := modeString(IOModeNone)
	if r != "0:000000" {
		t.Error("IOModeNone", r)
	}

	r = modeString(IOModeInput)
	if r != "1:000001" {
		t.Error("IOModeInput", r)
	}

	r = modeString(IOModeOutput)
	if r != "2:000010" {
		t.Error("IOModeOutput", r)
	}

	r = modeString(IOModeError)
	if r != "4:000100" {
		t.Error("IOModeError", r)
	}

	r = modeString(IOModeCreate)
	if r != "8:001000" {
		t.Error("IOModeCreate", r)
	}

	r = modeString(IOModeAppend)
	if r != "16:010000" {
		t.Error("IOModeCreate", r)
	}
}

func modeString(mode IOMode) string {
	return fmt.Sprintf("%d:%06b", mode, mode)
}

func TestNewCopy(t *testing.T) {

	ios := newIOSettings()

	ios.addFilePath(OutputIO, "test123", IOModeOutput)

	copy := ios.newCopy()

	fd := copy.getFileDetails(OutputIO)

	if fd == nil || fd.filePath != "test123" || fd.ioMode != IOModeOutput {
		t.Error("Copy missing correct data", fd)
	}
}

func TestDuplicateErrorsOnMissing(t *testing.T) {

	ios := newIOSettings()

	if err := ios.duplicate(OutputIO, ErrorIO); err == nil {
		t.Error("duplicate no error on missing source")
	}
}

func TestReadFileFailsWrongType(t *testing.T) {

	ios := newIOSettings()
	fd := ios.addFilePath(OutputIO, "context.go", IOModeOutput)

	if _, err := fd.ReadFile(); err == nil {
		t.Error("No error on output read")
	}
}

func TestReadFileSucceedsForIInput(t *testing.T) {

	ios := newIOSettings()
	fd := ios.addFilePath(InputIO, "context.go", IOModeInput)

	if b, err := fd.ReadFile(); err != nil {
		t.Error("Error on output read", err)
	} else if len(b) < 200 {
		t.Error("Error context.go too small when read", len(b))
	}
}

func TestOpenInputFailsWrongType(t *testing.T) {

	ios := newIOSettings()
	fd := ios.addFilePath(InputIO, "context.go", IOModeOutput)

	if _, err := fd.OpenInput(); err == nil {
		t.Error("No error on output read")
	}
}

func TestOpenInputSucceedsForIInput(t *testing.T) {

	ios := newIOSettings()
	fd := ios.addFilePath(InputIO, "context.go", IOModeInput)

	if f, err := fd.OpenInput(); err != nil {
		t.Error("Error on output read", err)
	} else {
		f.Close()
	}
}

func TestOpenOutput(t *testing.T) {

	ios := newIOSettings()
	fd := ios.addFilePath(InputIO, "dummy.txt", IOModeInput)

	if _, err := fd.OpenOutput(); err == nil {
		t.Error("no error  opening input as output")
	}

	fd = ios.addFilePath(InputIO, "dummy.txt", IOModeOutput)
	if _, err := fd.OpenOutput(); err == nil {
		t.Error("no error opening output no mode")
	}
}
