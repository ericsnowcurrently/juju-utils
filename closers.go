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

// CloseAll calls each of the provided closers. If any of them returns
// an errors then the error is stored and the handler (if any) is
// called. The resulting error will be an errors.BulkError.
func CloseAll(handleErr func(string, error), closers ...io.Closer) error {
	var ids []string
	for i := range closers {
		ids = append(ids, fmt.Sprintf("closer %d/%d", i+1, len(closers)))
	}
	err, setError := errors.NewBulkError(ids...)

	for i, closer := range closers {
		if err := closer.Close(); err != nil {
			id := ids[i]
			setError(id, errors.Trace(err))
			if handleErr == nil {
				continue
			}
			handleErr(id, err)
		}
	}

	if !err.NoErrors() {
		return errors.Trace(err)
	}
	return nil
}
