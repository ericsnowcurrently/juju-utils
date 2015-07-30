// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fs

import (
	"github.com/juju/errors"

	"github.com/juju/utils/filepath"
)

// Path represents a path in the filesystem.
type Path struct {
	raw      string
	renderer filepath.Renderer
}

// NewPath returns a new Path for the given path parts.
func NewPath(parts ...string) (*Path, error) {
	renderer, err := filepath.NewRenderer("")
	if err != nil {
		return nil, errors.Trace(err)
	}
	raw := renderer.Join(parts...)
	p := &Path{
		raw:      raw,
		renderer: renderer,
	}
	return p, nil
}

// String returns the path as a string.
func (p Path) String() string {
	return p.raw
}

func (p Path) resolve(parts ...string) Path {
	resolved := p.renderer.Join(append([]string{p.raw}, parts...)...)
	return Path{raw: resolved}
}

// VolumeName is equivalent to filepath.VolumeName.
func (p Path) VolumeName() string {
	return p.renderer.VolumeName(p.raw)
}

// Dir is equivalent to filepath.Dir.
func (p Path) Dir() Path {
	dirname := p.renderer.Dir(p.raw)
	return Path{raw: dirname}
}

// Base is equivalent to filepath.Base.
func (p Path) Base() string {
	return p.renderer.Base(p.raw)
}

func (p Path) suffix() string {
	return p.renderer.Ext(p.raw)
}

// Clean is equivalent to filepath.Clean.
func (p Path) Clean() Path {
	cleaned := p.renderer.Clean(p.raw)
	return Path{raw: cleaned}
}
