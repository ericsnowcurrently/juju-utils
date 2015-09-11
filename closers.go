// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"fmt"
	"io"

	"github.com/juju/errors"
)

type wrappingCloser struct {
	closeFunc func() error
}

// NewCloser wraps the provided "close" function in a naive io.Closer.
func NewCloser(closeFunc func() error) io.Closer {
	return &wrappingCloser{
		closeFunc: closeFunc,
	}
}

// Close implements io.Closer.
func (wc wrappingCloser) Close() error {
	return wc.closeFunc()
}

// Multicloser is an io.Closer that wraps multiple other closers. It
// closes them in the same order in which they were added.
type MultiCloser interface {
	io.Closer
	// AddClosers adds the provided closers to the set of closers.
	AddClosers(...io.Closer)
}

type multiCloser struct {
	closers    []io.Closer
	errHandler func(string, error)
}

// NewMultiCloser creates a new MultiCloser.
func NewMultiCloser(errHandler func(string, error)) MultiCloser {
	return &multiCloser{
		errHandler: errHandler,
	}
}

// AddCloser implements MultiCloser.
func (mc *multiCloser) AddClosers(closers ...io.Closer) {
	mc.closers = append(mc.closers, closers...)
}

// Close implements MultiCloser.
func (mc multiCloser) Close() error {
	closers := mc.closers
	var ids []string
	for i := range closers {
		ids = append(ids, fmt.Sprintf("#%d", i))
	}
	err, setError := errors.NewBulkError(ids...)

	for i, closer := range closers {
		if err := closer.Close(); err != nil {
			id := ids[i]
			setError(id, errors.Trace(err))
			if mc.errHandler == nil {
				// TODO(ericsnow) Fail by default?
				continue
			}
			mc.errHandler(id, err)
		}
	}

	if !err.NoErrors() {
		return errors.Trace(err)
	}
	return nil
}
