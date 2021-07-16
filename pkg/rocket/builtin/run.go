package builtin

import (
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
	// Run task is used to run a specific task.
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

		// LogOutput if true will cause output to be logged rather than going to go to std output.
		// If an output file is specified it will be used instead.
		LogOutput bool `mapstructure:"logStdOut"`

		// DirectError when true causes the commands std error output to go direct to running processes std error
		// When DirectError is false std error output is logged.
		DirectError bool `mapstructure:"directStdErr"`
	}

	runType struct{}
)

func (runType) Type() string {
	return "run"
}

func (runType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	runCfg := &Run{}

	if err := mapstructure.Decode(task.Definition, runCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	// Get the command line
	commandLine, err := getCommandLine(ctx, capComm, runCfg)
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

		// Run command
		var runExitCode int
		err := runCmd(execCtx, capComm, runCfg, cmd)
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

func runCmd(ctx context.Context, capComm *rocket.CapComm, runCfg *Run, cmd *exec.Cmd) error {
	pipeFuncs, cf, err := setupRedirect(capComm, cmd, runCfg)
	defer cf.Close() // close any files we opened in redirect
	if err != nil {
		return err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return err
	}

	// setup signal handler and close on exit
	signalHandlerDoneChannel := startProcessSignalHandlee(ctx, cmd)
	defer close(signalHandlerDoneChannel)

	// Handle pipes
	channels := make([]chan struct{}, 0, len(pipeFuncs))
	for _, fn := range pipeFuncs {
		channels = append(channels, fn())
	}

	// Wait for channels to close, meaning pipe has shut
	for _, ch := range channels {
		<-ch
	}

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
	glob := filepath.Glob
	if !runCfg.GlobArgs {
		glob = nil
	}

	return cliparse.NewParse().WithGlob(glob).Parse(cmd, args...)
}

// RegisterAll all built in task types with the passed mission control.
func RegisterAll(mc rocket.MissionControl) {
	mc.RegisterTaskTypes(templateType{}, runType{}, cleanerType{}, fetchType{}, mkDirType{})
}

func init() {
	rocket.Default().RegisterTaskTypes(runType{})
}
