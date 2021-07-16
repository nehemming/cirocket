package builtin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
)

type (
	// Fetch task is used get one or more web resources.
	Fetch struct {
		Resources []FetchResource `mapstructure:"resources"`
		Log       bool            `mapstructure:"log"`
		Timeout   *int            `mapstructure:"timeout"`
	}

	FetchResource struct {
		URL    string `mapstructure:"url"`
		Output string `mapstructure:"output"`
	}

	fetchType struct{}

	urlFileTuple struct {
		url string
		out string
	}
)

func (fetchType) Type() string {
	return "fetch"
}

func (fetchType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	fetchCfg := &Fetch{}

	if err := mapstructure.Decode(task.Definition, fetchCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	timeOut, err := getTimeOut(fetchCfg.Timeout)
	if err != nil {
		return nil, err
	}

	resources, err := getResourcesList(ctx, capComm, fetchCfg.Resources)
	if err != nil {
		return nil, err
	}

	fn := func(execCtx context.Context) error {
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
		var err error
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
		for _, res := range resources {
			// Check if exit requested
			if ctx.Err() != nil {
				return nil
			}

			// Run in parallel
			wg.Add(1)
			go func(working urlFileTuple) {
				defer wg.Done()
				if err := fetchResource(fetchCtx, working, timeOut, fetchCfg.Log); err != nil {
					errCh <- err
				}
			}(res)
		}

		// Wait for all requests to be done
		wg.Wait()

		// Close the error channel as no more errors
		close(errCh)

		// Wait for err change to close
		<-doneCh
		return err
	}

	return fn, nil
}

func getResourcesList(ctx context.Context, capComm *rocket.CapComm, list []FetchResource) ([]urlFileTuple, error) {
	resources := make([]urlFileTuple, 0, len(list))
	for index, res := range list {
		tup := urlFileTuple{}
		if out, err := capComm.ExpandString(ctx, "out", res.Output); err != nil {
			return nil, errors.Wrapf(err, "expanding output %d", index)
		} else if out == "" {
			return nil, fmt.Errorf("output %d is blank", index)
		} else {
			tup.out = out
		}

		if url, err := capComm.ExpandString(ctx, "url", res.URL); err != nil {
			return nil, errors.Wrapf(err, "expanding url %d", index)
		} else if url == "" {
			return nil, fmt.Errorf("url %d is blank", index)
		} else {
			tup.url = url
		}

		resources = append(resources, tup)
	}

	return resources, nil
}

func getTimeOut(cfgTimeout *int) (time.Duration, error) {
	var timeout time.Duration
	if cfgTimeout != nil {
		timeout = time.Second * time.Duration(*cfgTimeout)
	} else {
		timeout = time.Second * time.Duration(30)
	}
	if timeout < time.Second {
		return timeout, fmt.Errorf("timeout %d is to short", timeout/time.Second)
	}

	return timeout, nil
}

func fetchResource(ctx context.Context, res urlFileTuple, timeOut time.Duration, log bool) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeOut)
	defer cancel()

	req, err := http.NewRequest("GET", res.url, nil)
	if err != nil {
		return errors.Wrapf(err, "creating request for %s", res.url)
	}

	resp, err := ctxhttp.Do(ctxTimeout, nil, req)
	if err != nil {
		return errors.Wrapf(err, "getting %s", res.url)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// Bad response
		return fmt.Errorf("response (%d) %s for %s", resp.StatusCode, resp.Status, res.url)
	}

	// Create the file
	out, err := os.Create(res.out)
	if err != nil {
		return errors.Wrapf(err, "creating %s", res.out)
	}
	defer out.Close()

	// Write the body to file
	bytes, err := io.Copy(out, resp.Body)
	if log {
		loggee.Infof("fetch %s => %s, %d bytes", res.url, res.out, bytes)
	}
	return err
}

func init() {
	rocket.Default().RegisterTaskTypes(fetchType{})
}
