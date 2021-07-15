package rocket

import (
	"context"
)

type runCtx string

const ctxKey = runCtx("capcomm")

// GetRunContext returns the run context
func GetCapCommContext(ctx context.Context) *CapComm {
	if capComm, ok := ctx.Value(ctxKey).(*CapComm); !ok {
		return nil
	} else {
		return capComm
	}
}

// NewContextWithCapComm creates a new context with capComm attached
func NewContextWithCapComm(ctx context.Context, capComm *CapComm) context.Context {
	return context.WithValue(ctx, ctxKey, capComm.Seal())
}
