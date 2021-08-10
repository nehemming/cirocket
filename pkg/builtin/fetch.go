package builtin

import (
	"context"
	"io"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// Fetch task is used get one or more web resources.
	Fetch struct {
		Resources []FetchResource `mapstructure:"resources"`
		Log       bool            `mapstructure:"log"`
	}

	// FetchResource defines the input source and output target runbooks for a fetch request.
	FetchResource struct {
		Source rocket.InputSpec  `mapstructure:"source"`
		Output rocket.OutputSpec `mapstructure:"output"`
	}

	fetchType struct{}

	fetchOp struct {
		source io.ReadCloser
		target io.WriteCloser
	}

	fetchOps []*fetchOp
)

func (fetchType) Type() string {
	return "fetch"
}

func (fetchType) Description() string {
	return "fetches url bases resources and makes a local copy."
}

func (fetchType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	fetchCfg := &Fetch{}

	if err := mapstructure.Decode(task.Definition, fetchCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) (err error) {
		ops, err := getFetchOpsFromResourceList(ctx, capComm, fetchCfg.Resources)
		if err != nil {
			return
		}
		defer ops.Close()

		// Create a context we can cancel
		fetchCtx, cancel := context.WithCancel(execCtx)
		defer cancel()

		// set up channels
		// doneCh is closed by the error handler to indicate all errors captured
		doneCh := make(chan struct{})

		// channel with errors
		errCh := make(chan error)

		// Wait group for all requests to end
		wg := sync.WaitGroup{}

		// Consume errCh getting any errors form requests
		// Errors will cancel any remaining open requests

		go func() {
			defer close(doneCh)
			for e := range errCh {
				if err == nil {
					// First error will go to caller
					err = e
				} else {
					// Log others
					loggee.Error(e.Error())
				}

				// Signal stop processing other requests
				cancel()
			}
		}()

		// Run over requests
		for i, res := range ops {
			// Check if exit requested
			if fetchCtx.Err() != nil {
				return nil
			}

			// Run in parallel
			wg.Add(1)
			go func(index int, fetch *fetchOp) {
				defer wg.Done()

				_, err := io.Copy(fetch.target, fetch.source)
				if err != nil {
					errCh <- err
				}
				// Close asap
				fetch.Close()
				ops[index] = nil
			}(i, res)
		}

		// Wait for all requests to be done
		wg.Wait()

		// Close the error channel as no more errors
		close(errCh)

		// Wait for err channel to close
		<-doneCh
		return err
	}

	return fn, nil
}

func getFetchOpsFromResourceList(ctx context.Context, capComm *rocket.CapComm, list []FetchResource) (fetchOps, error) {
	resources := make(fetchOps, 0, len(list))
	for index, resource := range list {
		op, err := getResource(ctx, capComm, resource)
		if err != nil {
			resources.Close()
			return nil, errors.Wrapf(err, "resource[%d]", index)
		}

		resources = append(resources, op)
	}

	return resources, nil
}

func getResource(ctx context.Context, capComm *rocket.CapComm, resource FetchResource) (*fetchOp, error) {
	srcRp, err := capComm.InputSpecToResourceProvider(ctx, resource.Source)
	if err != nil {
		return nil, errors.Wrap(err, "source")
	}

	outRp, err := capComm.OutputSpecToResourceProvider(ctx, resource.Output)
	if err != nil {
		return nil, errors.Wrap(err, "output")
	}

	reader, err := srcRp.OpenRead(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "source")
	}

	writer, err := outRp.OpenWrite(ctx)
	if err != nil {
		reader.Close()
		return nil, errors.Wrap(err, "output")
	}

	return &fetchOp{
		source: reader,
		target: writer,
	}, nil
}

func (op fetchOp) Close() {
	_ = op.source.Close()
	_ = op.target.Close()
}

func (ops fetchOps) Close() {
	for _, op := range ops {
		if op != nil {
			op.Close()
		}
	}
}

func init() {
	rocket.Default().RegisterTaskTypes(fetchType{})
}
