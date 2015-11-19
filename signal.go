// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"os"

	"github.com/juju/errors"
)

// Signaler exposes the supported OS signals (see os.Signal).
type Signaler interface {
	Interrupter
	Killer
}

// Interrupter exposes the functionality to interrupt something.
type Interrupter interface {
	// Interrupt sends an interrupt signal.
	Interrupt() error
}

func interrupt(raw interface{}) (bool, error) {
	v, ok := raw.(Interrupter)
	if !ok {
		return false, nil
	}
	if err := v.Interrupt(); err != nil {
		return true, errors.Trace(err)
	}
	return true, nil
}

// Killer exposes the functionality to kill something.
type Killer interface {
	// Kill causes value to end immediately.
	Kill() error
}

func kill(raw interface{}) (bool, error) {
	v, ok := raw.(Killer)
	if !ok {
		return false, nil
	}
	if err := v.Kill(); err != nil {
		return true, errors.Trace(err)
	}
	return true, nil
}

// KillIfSupported calls Kill() on the provided value if it has the method.
func KillIfSupported(raw interface{}) error {
	if _, err := kill(raw); err != nil {
		return errors.Trace(err)
	}
	return nil
}

var recognizedSignals = map[os.Signal]func(interface{}) (bool, error){
	os.Interrupt: interrupt,
	os.Kill:      kill,
}

// Signal sends the OS signal to the given value. If it supports the
// requested signal then true and the error from the operation are
// returned. If the signal is not recognized then false and a
// errors.NotSupported are returned. Otherwise false and a nil error
// are returned.
func Signal(raw interface{}, sig os.Signal) (bool, error) {
	supported, err := signal(raw, sig, recognizedSignals)
	if err != nil {
		return supported, errors.Trace(err)
	}
	return supported, nil
}

func signal(raw interface{}, sig os.Signal, recognized map[os.Signal]func(interface{}) (bool, error)) (bool, error) {
	f, ok := recognizedSignals[sig]
	if !ok {
		return false, errors.NotSupportedf("signal %v", sig)
	}

	supported, err := f(raw)
	if err != nil {
		return supported, errors.Trace(err)
	}
	return supported, nil
}
