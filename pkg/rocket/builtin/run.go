package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/cliparse"
	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	// Run task is used to run a specific task
	Run struct {

		// Command to execute.  This is a raw executed command and is not
		// wrapped ina shell, as such globing etc may not work as expected.
		// Command is subject to string expansion
		Command string `mapstructure:"command"`

		// List of arguments, subject to string expansion
		Args []string `mapstructure:"args"`

		// GlobArgs specifies if arguments should be glob expanded prior
		// to passing the program.  If true *.go would ne expanded as a arg per matching file
		GlobArgs bool `mapstructure:"glob"`

		// Redirect handles input and output redirection
		rocket.Redirection `mapstructure:",squash"`

		// LogOutput if true will cause output to be logged rather than going to go to std output.
		// If an output file is specified it will be used instead
		LogOutput bool `mapstructure:"logStdOut"`

		// DirectError when true causes the commands std error output to go direct to running processes std error
		// When DirectError is false std error output is logged
		DirectError bool `mapstructure:"directStdErr"`
	}

	runType struct{}

	closeFiles []*os.File

	loggerFunc func() chan struct{}
)

func (runType) Type() string {
	return "run"
}

func (runType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {

	runCfg := &Run{}

	if err := mapstructure.Decode(task.Definition, runCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

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

	fn := func(execCtx context.Context) error {

		// Setup command
		cmd := exec.Command(commandLine.ProgramPath, commandLine.Args...)

		cmd.Env = capComm.GetExecEnv()

		// Check not cancelled
		if execCtx.Err() != nil {
			return nil
		}

		logFuncs, cf, err := setupRedirect(capComm, cmd, runCfg)

		defer cf.Close()
		if err != nil {
			return err
		}

		done := make(chan struct{})

		go func() {
			// Wait on any exit request
			select {
			case <-execCtx.Done():
				if cmd.Process != nil {
					cmd.Process.Signal(os.Interrupt)
				}
			case <-done:
				return
			}
		}()

		// Start
		if err := cmd.Start(); err != nil {
			return err
		}

		// Fire Log funcs
		channels := make([]chan struct{}, 0, len(logFuncs))
		for _, fn := range logFuncs {
			channels = append(channels, fn())
		}

		// Wait for channels to close, meaning pipe has shut
		for _, ch := range channels {
			<-ch
		}

		// Run command
		var runExitCode int
		if err = cmd.Wait(); err != nil {
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
	glob := filepath.Glob
	if !runCfg.GlobArgs {
		glob = nil
	}

	return cliparse.NewParse().WithGlob(glob).Parse(cmd, args...)
}

func setupRedirect(capComm *rocket.CapComm, cmd *exec.Cmd, runCfg *Run) ([]loggerFunc, closeFiles, error) {

	logFuncs := make([]loggerFunc, 0, 3)
	cf := make(closeFiles, 0, 3)
	var isOutAndErrMerged bool

	// Handle input
	inFd := capComm.GetFileDetails(rocket.InputIO)
	if inFd != nil {
		inFile, err := inFd.OpenInput()
		if err != nil {
			return logFuncs, cf, errors.Wrap(err, string(rocket.InputIO))
		}
		cf = append(cf, inFile)
		cmd.Stdin = inFile
	} else {
		cmd.Stdin = os.Stdin
	}

	// Handle output
	outFd := capComm.GetFileDetails(rocket.OutputIO)
	if outFd != nil {
		outFile, err := outFd.OpenOutput()
		if err != nil {
			return logFuncs, cf, errors.Wrap(err, string(rocket.OutputIO))
		}
		cf = append(cf, outFile)
		cmd.Stdout = outFile

		// Check if error is merged
		if outFd.InMode(rocket.IOModeError) {
			isOutAndErrMerged = true
			cmd.Stderr = outFile
		}
	} else if runCfg.LogOutput {

		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return logFuncs, cf, errors.Wrap(err, string(rocket.OutputIO))
		}

		fn := func() chan struct{} {

			ch := make(chan struct{})

			go func() {
				defer close(ch)
				in := bufio.NewScanner(outPipe)

				for in.Scan() {
					loggee.Info(in.Text()) // write each line to your log, or anything you need
				}
			}()

			return ch
		}

		logFuncs = append(logFuncs, fn)

	} else {
		// redirect to std out
		cmd.Stdout = os.Stdout
	}

	if !isOutAndErrMerged {

		errFd := capComm.GetFileDetails(rocket.ErrorIO)
		if errFd != nil {
			errFile, err := errFd.OpenOutput()
			if err != nil {
				return logFuncs, cf, errors.Wrap(err, string(rocket.ErrorIO))
			}
			cf = append(cf, errFile)
			cmd.Stderr = errFile
		} else if !runCfg.DirectError {

			errPipe, err := cmd.StderrPipe()
			if err != nil {
				return logFuncs, cf, errors.Wrap(err, string(rocket.OutputIO))
			}

			fn := func() chan struct{} {

				ch := make(chan struct{})

				go func() {
					defer close(ch)
					in := bufio.NewScanner(errPipe)

					for in.Scan() {
						loggee.Warn(in.Text()) // write each line to your log, or anything you need
					}
				}()

				return ch
			}

			logFuncs = append(logFuncs, fn)

		} else {
			cmd.Stderr = os.Stderr
		}
	}

	return logFuncs, cf, nil
}

// RegisterAll all built in task types with the passed mission control
func RegisterAll(mc rocket.MissionControl) {
	mc.RegisterTaskTypes(templateType{}, runType{}, cleanerType{}, fetchType{}, mkDirType{})
}

func init() {
	rocket.Default().RegisterTaskTypes(runType{})
}

func (cf closeFiles) Close() {
	for _, c := range cf {
		c.Close()
	}
}
