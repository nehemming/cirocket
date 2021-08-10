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

package builtin

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/cliparse"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// Run task is used to execute a specific command line program.
	Run struct {

		// Command to execute.  This is a raw executed command and is not
		// wrapped ina shell, as such globing etc may not work as expected.
		// Command is subject to string expansion.
		Command string `mapstructure:"command"`

		// List of arguments, subject to string expansion.
		Args []string `mapstructure:"args"`

		// GlobArgs specifies if arguments should be glob expanded prior
		// to passing the program.  If true *.go would ne expanded as a arg per matching file.
		GlobArgs bool `mapstructure:"glob"`

		// Redirect handles input and output redirection.
		rocket.Redirection `mapstructure:",squash"`
	}

	runType struct{}
)

func (runType) Type() string {
	return "run"
}

func (runType) Description() string {
	return "executes a program and awaits its response."
}

func (runType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	runCfg := &Run{}

	if err := mapstructure.Decode(task.Definition, runCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(execCtx context.Context) error {
		// Get the command line
		commandLine, err := getCommandLine(ctx, capComm, runCfg)
		if err != nil {
			return err
		}

		// Setup command
		cmd := exec.Command(commandLine.ProgramPath, commandLine.Args...)
		cmd.Env = capComm.GetExecEnv()

		// Check not cancelled
		if execCtx.Err() != nil {
			return nil
		}

		// Run command
		var runExitCode int
		err = runCmd(execCtx, capComm, cmd)
		if err != nil {
			// Issue caught
			if exitError, ok := err.(*exec.ExitError); ok {
				runExitCode = exitError.ExitCode()
			} else {
				return err
			}
		}

		if runExitCode != 0 {
			// Process failed
			return fmt.Errorf("process %s exit code %d", commandLine.ProgramPath, runExitCode)
		}

		return nil
	}

	return fn, nil
}

func getCommandLine(ctx context.Context, capComm *rocket.CapComm, runCfg *Run) (*cliparse.Commandline, error) {
	// Expand redirect settings into cap Comm
	if err := capComm.AttachRedirect(ctx, runCfg.Redirection); err != nil {
		return nil, errors.Wrap(err, "expanding redirection settings")
	}

	commandLine, err := parseCommandLine(ctx, capComm, runCfg)
	if err != nil {
		return nil, err
	}

	// Validate command can be found
	_, err = exec.LookPath(commandLine.ProgramPath)
	if err != nil {
		return nil, err
	}

	return commandLine, nil
}

func startProcessSignalHandlee(ctx context.Context, cmd *exec.Cmd) chan struct{} {
	done := make(chan struct{})

	go func() {
		// Wait on any exit request
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				err := cmd.Process.Signal(os.Interrupt)
				if err != nil {
					loggee.Warnf("run signal error: %s", err)
				}
			}
		case <-done:
			return
		}
	}()

	return done
}

func runCmd(ctx context.Context, capComm *rocket.CapComm, cmd *exec.Cmd) error {
	inputResource := capComm.GetResource(rocket.InputIO)
	outputResource := capComm.GetResource(rocket.OutputIO)
	errorResource := capComm.GetResource(rocket.ErrorIO)

	if inputResource != nil {
		stdIn, err := inputResource.OpenRead(ctx)
		if err != nil {
			return errors.Wrap(err, "input")
		}
		cmd.Stdin = stdIn
		defer stdIn.Close()
	}

	if outputResource != nil {
		stdOut, err := outputResource.OpenWrite(ctx)
		if err != nil {
			return errors.Wrap(err, "output")
		}
		cmd.Stdout = stdOut
		defer stdOut.Close()
	}

	if errorResource != nil {
		stdErr, err := errorResource.OpenWrite(ctx)
		if err != nil {
			return errors.Wrap(err, "error")
		}
		cmd.Stderr = stdErr
		defer stdErr.Close()
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return err
	}

	// setup signal handler and close on exit
	signalHandlerDoneChannel := startProcessSignalHandlee(ctx, cmd)
	defer close(signalHandlerDoneChannel)

	// Wait for process exit
	return cmd.Wait()
}

func parseCommandLine(ctx context.Context, capComm *rocket.CapComm, runCfg *Run) (*cliparse.Commandline, error) {
	cmd, err := capComm.ExpandString(ctx, "command", runCfg.Command)
	if err != nil {
		return nil, errors.Wrap(err, "parsing command line")
	}

	args := make([]string, 0)
	for index, a := range runCfg.Args {
		arg, err := capComm.ExpandString(ctx, "arg", a)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing arg %d", index)
		}
		args = append(args, arg)
	}

	// Should we glob
	glob := recursiveGlob
	if !runCfg.GlobArgs {
		glob = nil
	}

	return cliparse.NewParse().WithGlob(glob).Parse(cmd, args...)
}

func init() {
	rocket.Default().RegisterTaskTypes(runType{})
}
