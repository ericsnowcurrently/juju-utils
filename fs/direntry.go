// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"os"

	"github.com/juju/errors"
)

// DirEntry represents a "file" in the filesystem.
type DirEntry struct {
	path Path
}

// NewDirEntry returns a new DirEntry for the given path.
func NewDirEntry(pathParts ...string) (*DirEntry, error) {
	path, err := NewPath(pathParts...)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var de DirEntry
	de.path = *path
	return &de, nil
}

// String returns the path as a string.
func (de DirEntry) String() string {
	return de.path.String()
}

// Exists returns whether or not the path exists.
func (de DirEntry) Exists() (bool, error) {
	_, err := de.Stat()
	if os.IsNotExist(errors.Cause(err)) {
		return false, nil
	}
	if err != nil {
		return false, errors.Trace(err)
	}
	return true, nil
}

// Stat returns "file" information about the path, if it exists.
func (de DirEntry) Stat() (os.FileInfo, error) {
	info, err := os.Stat(de.path.String())
	if err != nil {
		return info, errors.Trace(err)
	}
	return info, nil
}

func (de DirEntry) lstat() (os.FileInfo, error) {
	info, err := os.Lstat(de.path.String())
	if err != nil {
		return info, errors.Trace(err)
	}
	return info, nil
}

// Remove deletes the "file" from the filesystem.
func (de DirEntry) Remove() error {
	if err := os.RemoveAll(de.path.String()); err != nil {
		return errors.Trace(err)
	}
	return nil
}
