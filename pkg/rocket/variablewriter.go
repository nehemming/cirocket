package rocket

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"
)

type variableWriter struct {
	data    bytes.Buffer
	name    string
	capComm *CapComm
}

func newVariableWriter(capComm *CapComm, name string) *variableWriter {
	return &variableWriter{
		capComm: capComm,
		name:    name,
	}
}

func (vw *variableWriter) OpenRead(ctx context.Context) (io.ReadCloser, error) {
	return nil, errors.New("variables cannot be read")
}

func (vw *variableWriter) OpenWrite(ctx context.Context) (io.WriteCloser, error) {
	return vw, nil
}

func (vw *variableWriter) Write(p []byte) (n int, err error) {
	return vw.data.Write(p)
}

func (vw *variableWriter) Close() error {
	vw.capComm.ExportVariable(vw.name, vw.data.String())
	return nil
}
