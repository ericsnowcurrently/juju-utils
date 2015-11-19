// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// exec provides utilities for executing commands through the OS.
package exec

import (
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("juju.utils.exec")

// Exec exposes the functionality of a command execution system.
type Exec interface {
	// FindExecutable looks for the named executable within the
	// execution system and returns a path that may be used in
	// CommandInfo.Path. If the executable cannot be found then
	// errors.NotFound is returned.
	FindExecutable(name string) (string, error)

	// Command returns a Command related to the system for the given info.
	Command(info CommandInfo) (Command, error)

	// TODO(ericsnow) Consider adding:
	//  - List() ([]Process, error)
	//  - Get(pid int) (Process, error)
}
