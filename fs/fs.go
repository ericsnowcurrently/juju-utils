// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/juju/errors"
)

// TODO(ericsnow) Implement a node table (a la inodes)?
// TODO(ericsnow) Support a "noclobber" option?
// TODO(ericsnow) Support default permissions and a perm mask?
// TODO(ericsnow) Add a path cache for search efficiency?

type FileSystem struct {
	root Dir
}

func (fs FileSystem) Root() Dir {
	return fs.root
}

func (fs FileSystem) resolve(name string) Node {
	// The caller will have already called path.Clean on the name.
	if !path.IsAbs(name) {
		return nil
	}

	var paths []string
	for name != "/" {
		name, basename := path.Split(name)
		paths = apend(paths, basename)
	}

	// Unroll from the root down to the leaf.
	parent := &fs.root
	var node Node = parent
	for i := len(paths) - 1; i >= 0; i-- {
		name := paths[i]
		var ok bool
		if node, ok = parent.Nodes[name]; !ok {
			return nil
		}
	}
	return node
}

func (fs *FileSystem) ResolveDir(dirname string) (name, *Dir) {
	name := path.Clean(name)

	node := fs.resolve(name)
	dir, _ := node.(*Dir)
	return name, dir
}

func (fs *FileSystem) ResolveFile(filename string) (string, *Dir, *File) {
	cleanName := path.Clean(name)

	var dirname, basename string
	if strings.HasSuffix(name, "/") {
		dirname, basename = cleanName, ""
		cleanName += "/"
	} else {
		dirname, basename := path.Split(cleanName)
	}

	dir := fs.ResolveDir(dirname)
	if dir == nil {
		return cleanName, nil, nil
	}

	return cleanName, dir, dir.File(basename)
}

func (fs *FileSystem) CreateFile(newname string) (*FileData, error) {
}

// Move disassociates the named file in the original directory and
// associates it with the new directory.
func (fs *FileSystem) Move(oldname, newname string) (Node, error) {
}

func (fs *FileSystem) Copy(name, newname string) (Node, error) {
}

func (fs *FileSystem) Remove(name string) (Node, error) {
}

func (fs *FileSystem) Link(name, newname string) error {
}

func (fs *FileSystem) Synlink(name, newname string) error {
}

func (fs *FileSystem) Touch(name string) error {
}

func (fs *FileSystem) Glob(pattern string) ([]Node, error) {
}

func (fs *FileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	dirname, dir := fs.ResolveDir(root)
	if dir == nil {
		// TODO(ericsnow) Return os.ErrNotExist?
		return nil
	}
	err := dir.Walk(dirname, walkFn)
	return errors.Trace(err)
}

func (fs *FileSystem) Attach(name string, fs *FileSystem) error {
}
