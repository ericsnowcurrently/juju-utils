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

	// AddCloser adds the identified closer to the set of closers.
	AddCloser(string, io.Closer)

	// AddClosers adds the provided closers to the set of closers.
	AddClosers(...io.Closer)
}

type multiCloser struct {
	ids        []string
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
func (mc *multiCloser) AddCloser(id string, closer io.Closer) {
	// TODO(ericsnow) This isn't thread-safe.
	if id == "" {
		id = fmt.Sprintf("#%d", len(mc.ids))
	}
	// TODO(ericsnow) Handle ID collisions?
	mc.ids = append(mc.ids, id)
	mc.closers = append(mc.closers, closer)
}

// AddCloser implements MultiCloser.
func (mc *multiCloser) AddClosers(closers ...io.Closer) {
	for _, closer := range closers {
		mc.AddCloser("", closer)
	}
}

// Close implements MultiCloser.
func (mc multiCloser) Close() error {
	// TODO(ericsnow) This isn't thread-safe.
	err, setError := errors.NewBulkError(mc.ids...)

	for i, closer := range mc.closers {
		if err := closer.Close(); err != nil {
			id := mc.ids[i]
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
