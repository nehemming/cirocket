package builtin

import (
	"bufio"
	"os"
	"os/exec"

	"github.com/nehemming/cirocket/pkg/loggee"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

type (
	closeFiles []*os.File

	pipeFuncs func() chan struct{}
)

func (cf closeFiles) Close() {
	for _, c := range cf {
		c.Close()
	}
}

type redirectWorker struct {
	pipeWorkers       []pipeFuncs
	fileClosers       closeFiles
	isOutAndErrMerged bool
}

func (redirect *redirectWorker) setUpStdIn(capComm *rocket.CapComm, cmd *exec.Cmd) error {
	// Handle input
	inFd := capComm.GetFileDetails(rocket.InputIO)
	if inFd != nil {
		inFile, err := inFd.OpenInput()
		if err != nil {
			return errors.Wrap(err, string(rocket.InputIO))
		}
		redirect.fileClosers = append(redirect.fileClosers, inFile)
		cmd.Stdin = inFile
	} else {
		cmd.Stdin = os.Stdin
	}

	return nil
}

func (redirect *redirectWorker) setUpStdOut(capComm *rocket.CapComm, cmd *exec.Cmd, logOutput bool) error {
	// Handle output
	outFd := capComm.GetFileDetails(rocket.OutputIO)
	if outFd != nil {
		outFile, err := outFd.OpenOutput()
		if err != nil {
			return errors.Wrap(err, string(rocket.OutputIO))
		}
		redirect.fileClosers = append(redirect.fileClosers, outFile)
		cmd.Stdout = outFile

		// Check if error is merged
		if outFd.InMode(rocket.IOModeError) {
			redirect.isOutAndErrMerged = true
			cmd.Stderr = outFile
		}
	} else if logOutput {
		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return errors.Wrap(err, string(rocket.OutputIO))
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

		redirect.pipeWorkers = append(redirect.pipeWorkers, fn)
	} else {
		// redirect to std out
		cmd.Stdout = os.Stdout
	}

	return nil
}

func (redirect *redirectWorker) setUpStdErr(capComm *rocket.CapComm, cmd *exec.Cmd, logOutput bool) error {
	errFd := capComm.GetFileDetails(rocket.ErrorIO)
	if errFd != nil {
		errFile, err := errFd.OpenOutput()
		if err != nil {
			return errors.Wrap(err, string(rocket.ErrorIO))
		}
		redirect.fileClosers = append(redirect.fileClosers, errFile)
		cmd.Stderr = errFile
	} else if logOutput {
		errPipe, err := cmd.StderrPipe()
		if err != nil {
			return errors.Wrap(err, string(rocket.OutputIO))
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

		redirect.pipeWorkers = append(redirect.pipeWorkers, fn)

	} else {
		cmd.Stderr = os.Stderr
	}

	return nil
}

func setupRedirect(capComm *rocket.CapComm, cmd *exec.Cmd, runCfg *Run) ([]pipeFuncs, closeFiles, error) {
	redirect := &redirectWorker{
		pipeWorkers: make([]pipeFuncs, 0, 3),
		fileClosers: make(closeFiles, 0, 3),
	}

	if err := redirect.setUpStdIn(capComm, cmd); err != nil {
		return redirect.pipeWorkers, redirect.fileClosers, err
	}

	if err := redirect.setUpStdOut(capComm, cmd, runCfg.LogOutput); err != nil {
		return redirect.pipeWorkers, redirect.fileClosers, err
	}

	if !redirect.isOutAndErrMerged {
		if err := redirect.setUpStdErr(capComm, cmd, !runCfg.DirectError); err != nil {
			return redirect.pipeWorkers, redirect.fileClosers, err
		}
	}

	return redirect.pipeWorkers, redirect.fileClosers, nil
}
