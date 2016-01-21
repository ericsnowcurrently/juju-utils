// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package pathhandler

import (
	"path"
	"path/filepath"
	"strings"
)

type TreeHandler struct {
	PathHandler

	RootPath string

	NormalizePath func(path string) (normalizedPath, sep string)
}

func NewTreeHandler(rootPath string, handler PathHandler) *TreeHandler {
	th := &TreeHandler{
		PathHandler: handler,
		RootPath:    rootPath,
		NormalizePath: func(pth string) (string, string) {
			normalized := filepath.ToSlash(pth) // ...just in case.
			normalized = path.Clean(normalized)
			return normalized, "/"
		},
	}
	return th
}

func (th *TreeHandler) ExtractPath(path string) string {
	// TODO(ericsnow) Normalize the root path first?
	path, sep := th.NormalizePath(path)
	path = strings.TrimPrefix(path, th.RootPath)
	return strings.TrimPrefix(path, sep)
}
