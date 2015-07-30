// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"github.com/juju/errors"
)

// Workspace is filesystem directory with cleanup.
type Workspace struct {
	Directory
	name   string
	closed bool
}

// NewWorkspace creates a new temporary workspace for the given name.
func NewWorkspace(name string) (*Workspace, error) {
	if name == "" {
		return nil, errors.Errorf("missing name")
	}

	rootdir, err := NewTempDirectory("workspace-" + name + "-")
	if err != nil {
		return nil, errors.Annotate(err, "while creating workspace dir")
	}

	ws := &Workspace{
		Directory: *rootdir,
		name:      name,
	}
	return ws, nil
}

// Close cleans up the workspace, deleting all files and directories.
func (ws Workspace) Close() error {
	if ws.closed {
		return nil
	}
	if err := ws.Remove(); err != nil {
		if exists, err := ws.Exists(); err == nil && !exists {
			ws.closed = true
		}
		return errors.Annotate(err, "while closing workspace")
	}
	ws.closed = true
	return nil
}
