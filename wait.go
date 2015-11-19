// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"reflect"
	"time"

	"github.com/juju/errors"
)

// Waiter represents something that waits.
type Waiter interface {
	// Wait for completion.
	Wait() error
}

// WaitForError waits for one of the provided channels to receive. The
// error (or nil) that it receives is returned.
func WaitForError(channels ...<-chan error) error {
	var cases []reflect.SelectCase
	for _, ch := range channels {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}
	_, val, recvOK := reflect.Select(cases)
	if !recvOK {
		// At least one of the channels was closed.
		return nil
	}
	err := val.Interface().(error)
	return errors.Trace(err)
}

type waitError struct {
	error
}

// WaitAbortable waits until the done channel or one of the abort
// channels receives. If the wait is aborted then false is returned.
// Otherwise true is returned.
func WaitAbortable(done <-chan error, abort ...<-chan error) (bool, error) {
	waitCh := make(chan error, 1)
	defer close(waitCh)
	go func() {
		err := <-done
		waitCh <- &waitError{err}
	}()

	err := WaitForError(append([]<-chan error{waitCh}, abort...)...)
	if err != nil {
		if waitErr, ok := err.(*waitError); ok {
			return true, errors.Trace(waitErr)
		}

		return false, errors.Trace(err)
	}

	return true, nil
}

// WaitTimedOut indicates that a wait timed out.
var WaitTimedOut = errors.New("timed out while waiting")

// WaitWithTimeout waits for an operation to finish. If the provided
// timeout channel receives before then, the process is killed.
//
// Combine this with clock.WallClock.After() to use a timeout duration.
func WaitWithTimeout(done <-chan error, timeoutCh <-chan time.Time) error {
	abortCh := make(chan error, 1)
	go func() {
		<-timeoutCh
		abortCh <- WaitTimedOut
	}()

	if _, err := WaitAbortable(done, abortCh); err != nil {
		return errors.Trace(err)
	}
	return nil
}
