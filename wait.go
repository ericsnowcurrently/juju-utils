// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"reflect"

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
