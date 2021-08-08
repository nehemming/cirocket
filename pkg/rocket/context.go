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
