package rocket

import (
	"context"
)

type runCtx string

const ctxKey = runCtx("capcomm")

// GetCapCommContext returns the capComm from the context.
func GetCapCommContext(ctx context.Context) *CapComm {
	capComm, ok := ctx.Value(ctxKey).(*CapComm)
	if !ok {
		return nil
	}
	return capComm
}

// NewContextWithCapComm creates a new context with capComm attached.
func NewContextWithCapComm(ctx context.Context, capComm *CapComm) context.Context {
	return context.WithValue(ctx, ctxKey, capComm.Seal())
}
