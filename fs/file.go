// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"io"
	"os"

	"github.com/juju/errors"
)

// File represents a regular file in the filesystem.
type File struct {
	DirEntry
}

func newFile(path Path) File {
	var f File
	f.path = path
	return f
}

// NewFile returns a new File for the given path.
func NewFile(pathParts ...string) (*File, error) {
	de, err := NewDirEntry(pathParts...)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var f File
	f.DirEntry = *de
	return &f, nil
}

// Create creates the file on the filesystem with the given file mode.
// The file is opened for read/write and returned. If the file already
// exists then an error is returned.
func (f File) Create(perms os.FileMode) (io.ReadWriteCloser, error) {
	mode := perms & os.ModePerm
	fd, err := os.OpenFile(f.String(), os.O_RDWR|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return fd, nil
}

// Open opens the file for read/write and returns it.
func (f File) Open() (io.ReadWriteCloser, error) {
	fd, err := os.Open(f.String())
	if err != nil {
		return nil, errors.Trace(err)
	}
	return fd, nil
}
