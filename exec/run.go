// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"bytes"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/juju/errors"

	"github.com/juju/utils"
)

// Run starts the provided command and waits for it to complete.
func Run(cmd Starter) (ProcessState, error) {
	process, err := cmd.Start()
	if err != nil {
		return nil, errors.Trace(err)
	}

	state, err := process.Wait()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return state, nil
}

// Output runs the command and returns its standard output.
func Output(cmd Command) ([]byte, error) {
	var buf bytes.Buffer
	stdio := Stdio{
		Out: &buf,
	}
	if err := cmd.SetStdio(stdio); err != nil {
		return nil, errors.Trace(err)
	}

	if _, err := Run(cmd); err != nil {
		return nil, errors.Trace(err)
	}

	return buf.Bytes(), nil
}

// CombinedOutput runs the command and returns its combined standard
// output and standard error.
func CombinedOutput(cmd Command) ([]byte, error) {
	var buf bytes.Buffer
	stdio := Stdio{
		Out: &buf,
		Err: &buf,
	}
	if err := cmd.SetStdio(stdio); err != nil {
		return nil, errors.Trace(err)
	}

	if _, err := Run(cmd); err != nil {
		return nil, errors.Trace(err)
	}

	return buf.Bytes(), nil
}

type waitError struct {
	error
}

// CommandTimedOut indicates that a command timed out.
var CommandTimedOut = errors.New("command timed out")

// WaitWithTimeout waits for the process to finish. If the provided
// channel receives before then, the process is killed.
func WaitWithTimeout(proc ProcessControl, timeoutCh <-chan time.Time) (ProcessState, error) {
	abortCh := make(chan error, 1)
	go func() {
		<-timeoutCh
		abortCh <- CommandTimedOut
	}()

	state, err := WaitAbortable(proc, abortCh)
	if err != nil {
		return state, errors.Trace(err)
	}
	return state, nil
}

// WaitWithTimeout waits for the process to finish. If the provided
// channels receive before then, the process is killed.
func WaitAbortable(proc ProcessControl, abortChans ...<-chan error) (ProcessState, error) {
	var state ProcessState

	done := make(chan error, 1)
	go func() {
		defer close(done)
		st, err := proc.Wait()
		state = st
		done <- &waitError{err}
	}()

	err := utils.WaitForError(append([]<-chan error{done}, abortChans...)...)
	if err != nil {
		if waitErr, ok := err.(*waitError); ok {
			return state, errors.Trace(waitErr)
		}

		// Abort!
		if err := proc.Kill(); err != nil {
			logger.Errorf("while killing process: %v", err)
		}
		<-done      // Wait for the command to finish.
		proc.Wait() // Finalize the command.
		return state, errors.Trace(err)
	}

	return state, nil
}

// TODO(ericsnow) Replace usage of RunCommands and RunParams with
// something more closely tied to Exec.

// RunCommands executes the Commands specified in the RunParams using
// powershell on windows, and '/bin/bash -s' on everything else,
// passing the commands through as stdin, and collecting
// stdout and stderr.  If a non-zero return code is returned, this is
// collected as the code for the response and this does not classify as an
// error.
func RunCommands(run RunParams) (*ExecResponse, error) {
	err := run.Run()
	if err != nil {
		return nil, err
	}
	return run.Wait()
}

// Parameters for RunCommands.  Commands contains one or more commands to be
// executed using bash or PowerShell.  If WorkingDir is set, this is passed
// through.  Similarly if the Environment is specified, this is used
// for executing the command.
type RunParams struct {
	Commands    string
	WorkingDir  string
	Environment []string

	tempDir string
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	ps      *osexec.Cmd
}

// Run sets up the command environment (environment variables, working dir)
// and starts the process. The commands are passed into bash on Linux machines
// and to powershell on Windows machines.
func (r *RunParams) Run() error {
	if runtime.GOOS == "windows" {
		r.Environment = mergeEnvironment(r.Environment)
	}

	tempDir, err := ioutil.TempDir("", "juju-exec")
	if err != nil {
		return err
	}

	shell, args, err := shellAndArgs(tempDir, r.Commands)
	if err != nil {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Warningf("failed to remove temporary directory: %v", err)
		}
		return err
	}

	r.ps = osexec.Command(shell, args...)
	if r.Environment != nil {
		r.ps.Env = r.Environment
	}
	if r.WorkingDir != "" {
		r.ps.Dir = r.WorkingDir
	}

	r.tempDir = tempDir
	r.stdout = &bytes.Buffer{}
	r.stderr = &bytes.Buffer{}

	r.ps.Stdout = r.stdout
	r.ps.Stderr = r.stderr

	return r.ps.Start()
}

// Process returns the *os.Process instance of the current running process
// This will allow us to kill the process if needed, or get more information
// on the process
func (r *RunParams) Process() *os.Process {
	if r.ps != nil && r.ps.Process != nil {
		return r.ps.Process
	}
	return nil
}

// Wait blocks until the process exits, and returns an ExecResponse type
// containing stdout, stderr and the return code of the process. If a non-zero
// return code is returned, this is collected as the code for the response and
// this does not classify as an error.
func (r *RunParams) Wait() (*ExecResponse, error) {
	var err error
	if r.ps == nil {
		return nil, errors.New("No process has been started yet")
	}
	err = r.ps.Wait()
	if err := os.RemoveAll(r.tempDir); err != nil {
		logger.Warningf("failed to remove temporary directory: %v", err)
	}

	result := &ExecResponse{
		Stdout: r.stdout.Bytes(),
		Stderr: r.stderr.Bytes(),
	}

	if ee, ok := err.(*osexec.ExitError); ok && err != nil {
		status := ee.ProcessState.Sys().(syscall.WaitStatus)
		if status.Exited() {
			// A non-zero return code isn't considered an error here.
			result.Code = status.ExitStatus()
			err = nil
		}
		logger.Infof("run result: %v", ee)
	}
	return result, err
}

// ExecResponse contains the return code and output generated by executing a
// command.
type ExecResponse struct {
	Code   int
	Stdout []byte
	Stderr []byte
}
