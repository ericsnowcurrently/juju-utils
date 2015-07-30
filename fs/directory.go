// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"io/ioutil"
	"os"

	"github.com/juju/errors"
)

// Directory represents a directory in the filesystem.
type Directory struct {
	DirEntry
}

func newDirectory(path Path) Directory {
	var dir Directory
	dir.path = path
	return dir
}

// NewDirectory returns a new directory for the given path.
func NewDirectory(pathParts ...string) (*Directory, error) {
	de, err := NewDirEntry(pathParts...)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var dir Directory
	dir.DirEntry = *de
	return &dir, nil
}

// NewTempDirectory creates a temporary directory with the given prefix.
func NewTempDirectory(prefix string) (*Directory, error) {
	path, err := ioutil.TempDir("", prefix)
	if err != nil {
		return nil, errors.Trace(err)
	}

	dir, err := NewDirectory(path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return dir, nil
}

// Resolve resolves the given relative path against the directory.
func (dir Directory) Resolve(relPath ...string) string {
	path := dir.path.resolve(relPath...)
	return path.String()
}

// Sub returns a new sub-Directory (relative to the directory) for
// the given name.
func (dir Directory) Sub(name string) Directory {
	path := dir.path.resolve(name)
	return newDirectory(path)
}

// File returns a new file for the given name with the path rooted at
// the directory.
func (dir Directory) File(name string) File {
	path := dir.path.resolve(name)
	return newFile(path)
}

// Create creates the directory on the filesystem. If it already exists
// then an error is returned.
func (dir Directory) Create(perms os.FileMode) error {
	mode := perms & os.ModePerm
	// TODO(ericsnow) Fail if the directory already exists.
	if err := os.MkdirAll(dir.String(), mode); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// List returns the contexts of the directory.
func (dir Directory) List() ([]File, []Directory, error) {
	infos, err := ioutil.ReadDir(dir.path.String())
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	var files []File
	var dirs []Directory
	for _, info := range infos {
		if info.Mode().IsDir() {
			dirs = append(dirs, dir.Sub(info.Name()))
		} else {
			files = append(files, dir.File(info.Name()))
		}
	}
	return files, dirs, nil
}
