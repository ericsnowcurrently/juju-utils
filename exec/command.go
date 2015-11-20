// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec

import (
	"fmt"
	"strings"

	"github.com/juju/errors"

	"github.com/juju/utils/filepath"
)

// Command exposes the functionality of a command.
//
// See os/exec.Cmd.
type Command interface {
	// Info returns the CommandInfo defining this Command.
	Info() CommandInfo

	StdioSetter
	Starter
}

// Starter describes a command that may be started.
type Starter interface {
	// Start starts execution of the command.
	Start() (Process, error)
}

// NewCommand returns a new Command for the given Exec and command.
func NewCommand(e Exec, path string, args ...string) (Command, error) {
	info := NewCommandInfo(path, args...)
	cmd, err := e.Command(info)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cmd, nil
}

// Cmd is a basic Command implementation.
type Cmd struct {
	CmdStdio
	Starter

	data CommandInfo
}

// Info implements Command.
func (c Cmd) Info() CommandInfo {
	return c.data
}

// CommandInfo holds the definition of a command's execution.
//
// See os/exec.Cmd.
type CommandInfo struct {
	// Path is the path to the command's executable.
	Path string

	// Args is the list of arguments to execute. Path must be Args[0].
	// If Args is not set then []string{Path} is used.
	Args []string

	Context
}

// NewCommandInfo returns a new CommandInfo for the given command. None
// of the command's context is set.
func NewCommandInfo(path string, args ...string) CommandInfo {
	// TODO(ericsnow) Call Exec.FindExecutable() for path?
	return CommandInfo{
		Path: path,
		Args: append([]string{path}, args...),
	}
}

// String returns a printable string for the info.
func (info CommandInfo) String() string {
	if info.Path == "" {
		return strings.Join(info.Args, " ")
	}
	if len(info.Args) == 0 {
		return info.Path
	}

	parts := append([]string{info.Path}, info.Args[1:]...)
	return strings.Join(parts, " ")
}

// TODO(ericsnow) Add Render(shell.Renderer) (string, error)
// It would incorporate at least the env.

// Validate ensures that the info is correct.
func (info CommandInfo) Validate() error {
	renderer, err := filepath.NewRenderer("")
	if err != nil {
		return errors.Trace(err)
	}
	if err := info.ValidateRendered(renderer); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ValidateRendered ensures that the info is correct. The renderer
// is used to check Path and the context.
func (info CommandInfo) ValidateRendered(renderer filepath.Renderer) error {
	if info.Path == "" {
		return errors.NewNotValid(nil, "missing command Path")
	}
	if len(info.Args) == 0 {
		return errors.NewNotValid(nil, "missing command Args (including command name)")
	}
	if renderer.Base(info.Path) != info.Args[0] {
		msg := fmt.Sprintf("command name mismatch: %q !-> %q", info.Path, info.Args[0])
		return errors.NewNotValid(nil, msg)
	}

	if err := info.Context.ValidateRendered(renderer); err != nil {
		return errors.Trace(err)
	}

	return nil
}
