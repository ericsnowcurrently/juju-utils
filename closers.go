// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils

import (
	"io"
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
