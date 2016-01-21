// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package pathhandler

import (
	"os"
)

type FileHandler interface {
	HandleDir(dirname string, finfo os.FileInfo) error
	HandleSymlink(path string, finfo os.FileInfo) error
	HandleRegular(filename string, finfo os.FileInfo) error
	HandleNamedPipe(path string, finfo os.FileInfo) error
	HandleSocket(path string, finfo os.FileInfo) error
	HandleDevice(path string, finfo os.FileInfo) error
}

type NopFileHandler struct{}

func (NopFileHandler) HandleDir(dirname string, finfo os.FileInfo) error      { return nil }
func (NopFileHandler) HandleSymlink(path string, finfo os.FileInfo) error     { return nil }
func (NopFileHandler) HandleRegular(filename string, finfo os.FileInfo) error { return nil }
func (NopFileHandler) HandleNamedPipe(path string, finfo os.FileInfo) error   { return nil }
func (NopFileHandler) HandleSocket(path string, finfo os.FileInfo) error      { return nil }
func (NopFileHandler) HandleDevice(path string, finfo os.FileInfo) error      { return nil }

func HandleFile(fh FileHandler, path string, finfo os.FileInfo) error {
	switch finfo.Mode() & os.ModeType {
	case os.ModeDir:
		return fh.HandleDir(path, finfo)
	case os.ModeSymlink:
		return fh.HandleSymlink(path, finfo)
	case os.ModeNamedPipe:
		return fh.HandleNamedPipe(path, finfo)
	case os.ModeSocket:
		return fh.HandleSocket(path, finfo)
	case os.ModeDevice:
		return fh.HandleDevice(path, finfo)
	default:
		return fh.HandleRegular(path, finfo)
	}
}
