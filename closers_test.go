// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

type wrappingCloserSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&wrappingCloserSuite{})

func (*wrappingCloserSuite) TestNewCloser(c *gc.C) {
	called := false
	closeFunc := func() error {
		called = true
		return nil
	}
	closer := utils.NewCloser(closeFunc)

	c.Check(closer, gc.NotNil)
	c.Check(called, jc.IsFalse)
}

func (*wrappingCloserSuite) TestCloseOkay(c *gc.C) {
	called := false
	closeFunc := func() error {
		called = true
		return nil
	}
	closer := utils.NewCloser(closeFunc)
	err := closer.Close()

	c.Check(err, jc.ErrorIsNil)
	c.Check(called, jc.IsTrue)
}

func (*wrappingCloserSuite) TestCloseError(c *gc.C) {
	failure := errors.Errorf("<failed>")
	called := false
	closeFunc := func() error {
		called = true
		return errors.Trace(failure)
	}
	closer := utils.NewCloser(closeFunc)
	err := closer.Close()

	c.Check(errors.Cause(err), gc.Equals, failure)
	c.Check(called, jc.IsTrue)
}
