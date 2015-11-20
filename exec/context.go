// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"github.com/juju/errors"

	"github.com/juju/utils/filepath"
)

// Context describes the context in which a command will run.
type Context struct {
	// Env is the list of environment variables to use. If Env is nil
	// then the current environment is used. If it is empty then
	// commands will run with no environment set.
	Env []string

	// Dir is the directory in which the command will be run. If omitted
	// then the current directory is used.
	Dir string

	// Stdio holds the stdio streams for the context.
	Stdio Stdio
}

// Validate checks the Context for correctness.
func (c Context) Validate() error {
	renderer, err := filepath.NewRenderer("")
	if err != nil {
		return errors.Trace(err)
	}
	if err := c.ValidateRendered(renderer); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ValidateRendered checks the Context for correctness.
func (c Context) ValidateRendered(renderer filepath.Renderer) error {
	// For now we don't check anything.
	return nil
}

// SetStdio sets the stdio this command will use. Nil values are
// ignored. Any non-nil value for which the corresponding current
// value is non-nil results in an error.
func (c Context) SetStdio(values Stdio) error {
	stdio, err := c.Stdio.WithInitial(values)
	if err != nil {
		return errors.Trace(err)
	}

	c.Stdio = stdio
	return nil
}
