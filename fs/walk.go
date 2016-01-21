// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juju/utils/pathhandler"
)

type DirWalker struct {
	PathHandler *pathhandler.FSPathHandler
}

func NewDirWalker(rootdir string, handler pathhandler.PathHandler) *DirWalker {
	fsPathHandler := pathhandler.NewFSPathHandler(rootdir, handler)
	fsPathHandler.FileHandler = &pathhandler.NopFileHandler{}
	dw := &DirWalker{
		PathHandler: fsPathHandler,
	}
	return dw
}

func (dw *DirWalker) Walk() error {
	return filepath.Walk(dw.PathHandler.RootPath, dw.HandleFile)
}

func (dw *DirWalker) HandleFile(path string, finfo os.FileInfo, err error) error {
	ph := dw.PathHandler.NewSinglePathHandler(finfo)
	return pathhandler.HandlePath(ph, path, err)
}

type TreeReader struct {
	DirWalker

	Tracker *FileTracker
}

func NewTreeReader(rootDir string, handler pathhandler.PathHandler) *TreeReader {
	tracker := NewFileTracker()
	dirWalker := NewDirWalker(rootDir, handler)
	dirWalker.FileHandler = tracker
	tr := &TreeReader{
		DirWalker: dirWalker,
		Tracker:   tracker,
	}
	return tr
}

type FileTracker struct {
	Regular map[string]os.FileInfo

	Irregular map[string]os.FileInfo
}

func NewFileTracker() *FileTracker {
	ft := &FileTracker{
		Regular:   make(map[string]os.FileInfo),
		Irregular: make(map[string]os.FileInfo),
	}
	return ft
}

func (ft *FileTracker) HandleRegular(filename string, finfo os.FileInfo) error {
	ft.Regular[filename] = finfo
	return nil
}

func (ft *FileTracker) HandleDir(dirname string, finfo os.FileInfo) error {
	ft.Irregular[dirname] = finfo
	return nil
}

func (ft *FileTracker) HandleSymlink(path string, finfo os.FileInfo) error {
	ft.Irregular[path] = finfo
	return nil
}

func (ft *FileTracker) HandleNamedPipe(path string, finfo os.FileInfo) error {
	ft.Irregular[path] = finfo
	return nil
}

func (ft *FileTracker) HandleSocket(path string, finfo os.FileInfo) error {
	ft.Irregular[path] = finfo
	return nil
}

func (ft *FileTracker) HandleDevice(path string, finfo os.FileInfo) error {
	ft.Irregular[path] = finfo
	return nil
}
