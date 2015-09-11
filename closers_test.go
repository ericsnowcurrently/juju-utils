// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"io"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var (
	_ = gc.Suite(&wrappingCloserSuite{})
	_ = gc.Suite(&closeAllSuite{})
)

type wrappingCloserSuite struct {
	testing.IsolationSuite
}

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

type closeAllSuite struct {
	testing.IsolationSuite
}

func (s *closeAllSuite) newCloser(name string, stub *testing.Stub) io.Closer {
	return utils.NewCloser(func() error {
		stub.AddCall("close-" + name)
		if err := stub.NextErr(); err != nil {
			return errors.Trace(err)
		}
		return nil
	})
}

func (s *closeAllSuite) TestOkay(c *gc.C) {
	var stub testing.Stub
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
	}
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)

	err := utils.CloseAll(errHandler, subCloserA, subCloserB)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "close-A", "close-B")
}

func (s *closeAllSuite) TestOkayNoHandler(c *gc.C) {
	var stub testing.Stub
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)

	err := utils.CloseAll(nil, subCloserA, subCloserB)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "close-A", "close-B")
}

func (s *closeAllSuite) TestNoClosers(c *gc.C) {
	var stub testing.Stub
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
	}

	err := utils.CloseAll(errHandler)
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCalls(c, nil)
}

func (s *closeAllSuite) TestOneError(c *gc.C) {
	var stub testing.Stub
	failure := errors.Errorf("<failure>")
	stub.SetErrors(nil, failure, nil, nil)
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
		stub.NextErr()
	}
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)
	subCloserC := s.newCloser("C", &stub)

	err := utils.CloseAll(errHandler, subCloserA, subCloserB, subCloserC)

	c.Check(err, gc.ErrorMatches, `1/3 items failed a bulk request: .*`)
	stub.CheckCallNames(c, "close-A", "close-B", "errHandler", "close-C")
}

func (s *closeAllSuite) TestMultiError(c *gc.C) {
	var stub testing.Stub
	failure := errors.Errorf("<failure>")
	stub.SetErrors(failure, nil, failure, nil, failure, nil)
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
		stub.NextErr()
	}
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)
	subCloserC := s.newCloser("C", &stub)

	err := utils.CloseAll(errHandler, subCloserA, subCloserB, subCloserC)

	c.Check(err, gc.ErrorMatches, `3/3 items failed a bulk request: .*`)
	stub.CheckCallNames(c,
		"close-A", "errHandler",
		"close-B", "errHandler",
		"close-C", "errHandler",
	)
}

func (s *closeAllSuite) TestErrorNoHandler(c *gc.C) {
	var stub testing.Stub
	failure := errors.Errorf("<failure>")
	stub.SetErrors(failure, nil, failure)
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)
	subCloserC := s.newCloser("C", &stub)

	err := utils.CloseAll(nil, subCloserA, subCloserB, subCloserC)

	c.Check(err, gc.ErrorMatches, `2/3 items failed a bulk request: .*`)
	stub.CheckCallNames(c, "close-A", "close-B", "close-C")
}
