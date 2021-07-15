package rocket

import (
	"fmt"
	"os"
)

const (
	InputIO  = NamedIO("input")
	OutputIO = NamedIO("output")
	ErrorIO  = NamedIO("error")
)

const (
	IOModeInput = IOMode(1 << iota)
	IOModeOutput
	IOModeError
	IOModeCreate
	IOModeAppend

	IOModeNone = IOMode(0)
)

type (
	IOMode uint32

	// NamedIO is a named io file
	NamedIO string

	// FileDetail represents file details
	FileDetail struct {
		filePath string
		ioMode   IOMode
		fileMode os.FileMode
	}

	// ioSettings is a collection of the file and IO settings used by a task
	ioSettings struct {
		files map[NamedIO]*FileDetail
	}
)

func newIOSettings() *ioSettings {
	return &ioSettings{
		files: make(map[NamedIO]*FileDetail),
	}
}

// Creates a new copy from the parent
func (ios *ioSettings) newCopy() *ioSettings {
	copy := &ioSettings{
		files: make(map[NamedIO]*FileDetail),
	}

	for k, v := range ios.files {
		copy.files[k] = v
	}

	return copy
}

func (ios *ioSettings) addFilePath(name NamedIO, filePath string, mode IOMode) *FileDetail {

	fd := &FileDetail{
		filePath: filePath,
		ioMode:   mode,
		fileMode: 0666,
	}

	ios.files[name] = fd

	return fd
}

func (ios *ioSettings) duplicate(from, to NamedIO) error {

	if f, ok := ios.files[from]; ok {
		ios.files[to] = f
		return nil
	}
	return fmt.Errorf("file type %s could not be found", from)
}

// getFileDetails returns the named file details or nil
func (ios *ioSettings) getFileDetails(name NamedIO) *FileDetail {
	return ios.files[name]
}

func (fd *FileDetail) FilePath() string {
	return fd.filePath
}

// Read the file
func (fd *FileDetail) ReadFile() ([]byte, error) {

	if (fd.ioMode & IOModeInput) == IOModeNone {
		return nil, fmt.Errorf("file type %s is nt an input file type", fd.filePath)
	}
	return os.ReadFile(fd.filePath)
}

// InMode returns true if the file in in the mode in question
func (fd *FileDetail) InMode(mode IOMode) bool {
	return (fd.ioMode & mode) == mode
}

// OpenOutput opens an output file
func (fd *FileDetail) OpenOutput() (*os.File, error) {

	if (fd.ioMode & (IOModeOutput | IOModeError)) == IOModeNone {
		return nil, fmt.Errorf("file type %s is nt an output file type", fd.filePath)
	}

	if (fd.ioMode & IOModeCreate) == IOModeCreate {
		return os.OpenFile(fd.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fd.fileMode)
	} else if (fd.ioMode & IOModeAppend) == IOModeAppend {
		return os.OpenFile(fd.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, fd.fileMode)
	} else {
		return nil, fmt.Errorf("file type %s does not specify create mode", fd.filePath)
	}
}

// Open for input
func (fd *FileDetail) OpenInput() (*os.File, error) {
	if (fd.ioMode & IOModeInput) == IOModeNone {
		return nil, fmt.Errorf("file type %s is nt an input file type", fd.filePath)
	}

	return os.OpenFile(fd.filePath, os.O_RDONLY, fd.fileMode)
}
