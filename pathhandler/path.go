// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package pathhandler

import (
	"github.com/juju/errors"
)

type PathHandler interface {
	NotifyPath(path string)
	ProcessPath(path string) error
	ProcessError(err error, path string) error
}

type NopPathHandler struct{}

func (NopPathHandler) NotifyPath(path string)                    {}
func (NopPathHandler) ProcessPath(path string) error             { return nil }
func (NopPathHandler) ProcessError(err error, path string) error { return nil }

func handlePath(ph PathHandler, path string, err error) error {
	// We can call this before processing the error
	// because it has no error return.
	ph.NotifyPath(path)

	if err := ph.ProcessError(err, path); err != nil {
		return errors.Trace(err)
	}

	if err := ph.ProcessPath(path); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type RawPathHandler interface {
	PathHandler
	ExtractPath(path string) string
}

type NopRawPathHandler struct {
	NopPathHandler
	//PathHandler
}

func (NopRawPathHandler) ExtractPath(path string) string { return path }

func HandlePath(ph RawPathHandler, path string, err error) error {
	relPath := ph.ExtractPath(path)
	if err := handlePath(ph, relPath, err); err != nil {
		return errors.Trace(err)
	}
	return nil
}
