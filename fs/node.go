// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package fs

import (
	"os"
	"time"

	"github.com/juju/errors"
)

const (
	NodeKindFile      = ""
	NodeKindDir       = "dir"
	NodeKindSymlink   = "symlink"
	NodeKindDevice    = "device"
	NodeKindSocket    = "socket"
	NodeKindNamedPipe = "namedpipe"
)

const (
	ModeUnknown = os.FileMode(0)
)

var (
	nodeKindModes = map[string]os.FileMode{
		NodeKindFile:      0,
		NodeKindDir:       os.ModeDir,
		NodeKindSymlink:   os.ModeSymlink,
		NodeKindDevice:    os.ModeDevice,
		NodeKindSocket:    os.ModeSocket,
		NodeKindNamedPipe: os.ModeNamedPipe,
	}
)

// Node represents a single filesystem node. It exposes all the
// characteristics that all types of "file" have in common.
//
// Consequently Node has all the information needed for the os.FileInfo
// inerface except for the file name. Unlike files (in all their types),
// nodes do not have a name. File names only exist in the context of
// directories, where each name in a direcory is associated with a
// single node. So one node may have many names but for each underlying
// file there is only one node in the filesystem.
type Node interface {
	// Info returns the info for the file.
	Info(name string) os.FileInfo

	// Touch updates the modification time of the node to the current time.
	Touch()

	// SetPermissions updates the permissions portion of the node's mode.
	SetPermissions()
}

// TODO(ericsnow) Add an ID field (for use in a filesystem)?
// TODO(ericsnow) Support CreationTime and AccessTime?

// NodeInfo is the implementation of Node.
type NodeInfo struct {
	// Size is the size of the node's content.
	Size int64

	// Mode is the node's permissions and other file mode values.
	Mode os.FileMode

	// ModTime is when the node was last modified.
	ModTime time.Time

	// CreationTime is when the node was created.
	CreationTime time.Time

	// AccessTime is when the node was last accessed.
	AccessTime time.Time

	// Owner is the UID of the use that owns the node, if applicable.
	Owner int

	// Group is the GID of the group that "owns" the node, if applicable.
	Group int
}

// newNode creates a new NodeInfo with the mode populated based on the
// provided node kind and all the timestamps set. If the node kind is
// not recognized then the mode will be set to ModeUnknown.
func newNode(kind string) NodeInfo {
	mode, ok := nodeKindModes[kind]
	if !ok {
		mode = ModeUnknown
	}
	info := NodeInfo{
		Mode: mode,
	}
	info.CreationTime = info.Touch()
	return info
}

// Info implements Node.
func (ni NodeInfo) Info(name string) os.FileInfo {
	return FileInfo{
		name: name,
		node: ni, // This makes a copy.
	}
}

// Touch implements Node.
func (ni *NodeInfo) Touch() time.Time {
	now := time.Now()
	ni.ModTime = now
	ni.AccessTime = now
}

// SetPermissions implements Node.
func (n *NodeInfo) SetPermissions(perm os.FileMode) {
	n.Mode = (^os.ModePerm & n.Mode) | (os.ModePerm & perm)
}

// fileInfo holds the information exposed by the os.FileInfo interface.
type FileInfo struct {
	name string
	node NodeInfo
}

// Name implements os.FileInfo.
func (fi FileInfo) Name() string {
	return fi.name
}

// Size implements os.FileInfo.
func (fi FileInfo) Size() int64 {
	return fi.node.Size
}

// Mode implements os.FileInfo.
func (fi FileInfo) Mode() os.FileMode {
	return fi.node.Mode
}

// ModTime implements os.FileInfo.
func (fi FileInfo) ModTime() time.Time {
	return fi.node.ModTime
}

// IsDir implements os.FileInfo.
func (fi FileInfo) IsDir() bool {
	return fi.node.Mode.IsDir()
}

// Sys implements os.FileInfo.
func (fi FileInfo) Sys() interface{} {
	// This is not implemented.
	return nil
}
