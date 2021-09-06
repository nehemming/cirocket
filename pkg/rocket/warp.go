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
	"os"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/resource"
	"github.com/pkg/errors"
)

// OperationHandler is a handler function that given a operations next function returns the function replacing it.
type OperationHandler func(next ExecuteFunc) ExecuteFunc

func (op *operation) AddHandler(handler OperationHandler) *operation {
	op.makeItSo = handler(op.makeItSo)

	return op
}

type concurrentErrors struct {
	list []error
	mu   sync.Mutex
}

func (ce *concurrentErrors) Add(err ...error) *concurrentErrors {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.list = append(ce.list, err...)
	return ce
}

func (ce *concurrentErrors) Error() error {
	if len(ce.list) == 0 {
		return nil
	}
	return &multierror.Error{Errors: ce.list}
}

func engage(ctx context.Context, operations operations, onFailStage *operation, log loggee.Logger) (err error) {
	//	Run mission
	forward := false
	for _, op := range operations {
		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}

		// running an op
		err = driveOp(ctx, op, log)
		forward = true
		if err != nil {
			break
		}
	}

	// if there was an error and something was done then apply reverse
	if err != nil && forward && onFailStage != nil {
		fullReverse(ctx, onFailStage.makeItSo, onFailStage.description, log)
	}

	return err
}

func warpEngines(ctx context.Context, ops operations, log loggee.Logger) error {
	// run each in parallel
	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := new(concurrentErrors)

	for _, op := range ops {
		if cancelCtx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(op *operation) {
			defer wg.Done()
			if cancelCtx.Err() != nil {
				return
			}
			if err := driveOp(cancelCtx, op, log); err != nil {
				cancel()

				// capture error
				errs.Add(err)
			}
		}(op)
	}

	wg.Wait()

	return errs.Error()
}

func engageWarpDrive(ops operations, log loggee.Logger) ExecuteFunc {
	return func(ctx context.Context) error {
		return log.Activity(ctx, func(ctx context.Context) error {
			return warpEngines(ctx, ops, log)
		})
	}
}

func impulseAhead(ops operations, dir string, log loggee.Logger) ExecuteFunc {
	return func(ctx context.Context) error {
		pop, err := swapDir(dir)
		if err != nil {
			return err
		}
		defer pop()

		for _, op := range ops {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := log.Activity(ctx, func(ctx context.Context) error {
				return driveOp(ctx, op, log)
			}); err != nil {
				return err
			}
		}

		return nil
	}
}

func driveOp(ctx context.Context, op *operation, log loggee.Logger) error {
	log.Info(op.description)

	if err := log.Activity(ctx, op.makeItSo); err != nil {
		if op.try {
			log.Warnf("try failed: %s", errors.Wrap(err, op.description))
		} else {
			if op.onFail != nil {
				fullReverse(ctx, op.onFail, op.description, log)
			}

			// report original error
			return errors.Wrap(err, op.description)
		}
	}

	return nil
}

func fullReverse(ctx context.Context, action ExecuteFunc, description string, log loggee.Logger) {
	// don't pass cancel context into fail action.
	fn := func(_ context.Context) error {
		return action(context.Background())
	}

	// Invoke failure fallback
	if err := log.Activity(ctx, fn); err != nil {
		log.Errorf("fail action failed: %s", errors.Wrap(err, description))
	}
}

// swapDir changes to the new directory and resurns a function to resore the current dir, or the functionreturns an error.
// If the restore function fails to restor the working dir it will panic.
func swapDir(dir string) (func(), error) {
	// return no op if no dir change requested
	if dir == "" {
		return func() {}, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// decode any urls
	u, err := resource.UltimateURL(dir)
	if err != nil {
		return nil, err
	}
	dir, err = resource.URLToPath(u)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	return func() {
		if e := os.Chdir(cwd); e != nil {
			panic(e)
		}
	}, nil
}
