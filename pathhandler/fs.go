// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package pathhandler

import (
	"os"
	"path/filepath"

	"github.com/juju/errors"
)

type FSPathHandler struct {
	*TreeHandler

	FileHandler FileHandler
}

func NewFSPathHandler(rootdir string, handler PathHandler) *FSPathHandler {
	treeHandler := NewTreeHandler(rootdir, handler)
	treeHandler.NormalizePath = func(path string) (string, string) {
		path = filepath.Clean(path)
		return path, string(os.PathSeparator)
	}
	ph := &FSPathHandler{
		TreeHandler: treeHandler,
		FileHandler: &NopFileHandler{},
	}
	return ph
}

func (ph FSPathHandler) ProcessFSPath(path string, finfo os.FileInfo) error {
	if err := HandleFile(ph.FileHandler, path, finfo); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (ph FSPathHandler) NewSinglePathHandler(finfo os.FileInfo) RawPathHandler {
	return &singleFSPathHandler{
		RawPathHandler: &ph,
		finfo:          finfo,
		processPath:    ph.ProcessFSPath,
	}
}

type singleFSPathHandler struct {
	RawPathHandler
	finfo       os.FileInfo
	processPath func(path string, finfo os.FileInfo) error
}

func (sfsph singleFSPathHandler) ProcessPath(path string) error {
	if err := sfsph.processPath(path, sfsph.finfo); err != nil {
		return errors.Trace(err)
	}
	return nil
}
