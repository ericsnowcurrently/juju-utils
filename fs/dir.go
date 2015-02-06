// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
    "os"

	"github.com/juju/errors"
)

// TODO(ericsnow) Resolve concurrency issues?

// ListSubdirectories extracts the names of all subdirectories of the
// specified directory and returns that list.
func ListSubdirectories(dirname string) ([]string, error) {
	return ListSubdirectoriesOp(dirname, &Ops{})
}

// ListSubdirectoriesOp extracts the names of all subdirectories of the
// specified directory and returns that list. The provided Operations
// is used to make the filesystem calls.
func ListSubdirectoriesOp(dirname string, fops Operations) ([]string, error) {
	entries, err := fops.ListDir(dirname)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var dirnames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirnames = append(dirnames, entry.Name())
	}
	return dirnames, nil
}

// At this abstraction layer we do not do any clobber checking.

type Dir struct{
    NodeInfo

    Nodes map[string]Node
}

func (d *Dir) List() []Node {
    return d.Nodes
}

func (d *Dir) SubDirs() []*Dir {
}

func (d *Dir) Walk(dirname string, walkFn filepath.WalkFunc) error {
}

// NewFile creates a new file and associated it with the provided name.
// If the file already exists then it is replaced with the new file.
func (d *Dir) NewFile(name string) *File {
    file := newFile()
    d.Nodes[name] = file
    return file
}

// NewSubDir creates a new directory and associated it with the provided
// name. If the name is already in use then the old node is replaced.
func (d *Dir) NewSubDir(name string) *Dir {
    dir := newDir()
    d.Nodes[name] = dir
    return dir
}

// File returns the file in the directory associated with the provided
// name. It is unchanged (except for any access time data). If the file
// does not exist the nil is returned.
func (d *Dir) File(name string) *File {
    file, _ := d.Nodes[name].(*File)
    return file
}

// SubDir returns the subdirectory associated with the provided name.
// It is unchanged (except for any access time data). If the file does
// not exist then nil is returned.
func (d *Dir) SubDir(name string) *Dir {
    dir, _ := d.Nodes[name].(*Dir)
    return dir
}

// Move disassociates the named file in the original directory and
// associates it with the new one.
func (d *Dir) Move(name, newname string, newdir *Dir) Node {
    if newname == "" {
        newname = name
    }
    if newdir == nil {
        newdir = d
    }

    node := d.Remove(name)
    if node != nil {
        return nil
    }

    newdir.Nodes[newname] = node
    return node
}

// TODO(ericsnow) Add full-featured copy capability (e.g. "cp") or just
// the minimal and let Filesystem handle the advanced capability.
//func (d *Dir) Copy(name, dir *Dir, directive ...string) *Dir {

func (d *Dir) Remove(name string) Node {
    node := d.Nodes[name]
    delete(d.Nodes, name)
}

func (d *Dir) () error {
}


func (fs *Filesystem) CreateFile(newname string) (*FileData, error) {
}

func (fs *Filesystem) Copy(name, newname string) (Node, error) {
}

func (fs *Filesystem) Remove(name string) (Node, error) {
}

func (fs *Filesystem) Link(name, newname string) error {
}

func (fs *Filesystem) Synlink(name, newname string) error {
}

func (fs *Filesystem) Attach(
