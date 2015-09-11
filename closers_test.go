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
	_ = gc.Suite(&multiCloserSuite{})
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

type multiCloserSuite struct {
	testing.IsolationSuite
}

func (s *multiCloserSuite) newCloser(name string, stub *testing.Stub) io.Closer {
	return utils.NewCloser(func() error {
		stub.AddCall("close-" + name)
		if err := stub.NextErr(); err != nil {
			return errors.Trace(err)
		}
		return nil
	})
}

func (s *multiCloserSuite) TestNewMultiCloserOkay(c *gc.C) {
	var stub testing.Stub
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
	}
	closer := utils.NewMultiCloser(errHandler)

	c.Check(closer, gc.NotNil)
	stub.CheckCalls(c, nil)
}

func (s *multiCloserSuite) TestNewMultiCloserNilErrHandler(c *gc.C) {
	closer := utils.NewMultiCloser(nil)

	c.Check(closer, gc.NotNil)
}

func (s *multiCloserSuite) TestAddClosersOkay(c *gc.C) {
	var stub testing.Stub
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)
	closer := utils.NewMultiCloser(nil)

	closer.AddClosers(subCloserA, subCloserB)
	err := closer.Close()
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "close-A", "close-B")
}

func (s *multiCloserSuite) TestAddClosersNone(c *gc.C) {
	closer := utils.NewMultiCloser(nil)

	closer.AddClosers()
	err := closer.Close()

	c.Check(err, jc.ErrorIsNil)
}

func (s *multiCloserSuite) TestCloseOkay(c *gc.C) {
	var stub testing.Stub
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
	}
	subCloserA := s.newCloser("A", &stub)
	subCloserB := s.newCloser("B", &stub)
	closer := utils.NewMultiCloser(errHandler)
	closer.AddClosers(subCloserA, subCloserB)

	err := closer.Close()
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "close-A", "close-B")
}

func (s *multiCloserSuite) TestCloseNoClosers(c *gc.C) {
	var stub testing.Stub
	errHandler := func(string, error) {
		stub.AddCall("errHandler")
	}
	closer := utils.NewMultiCloser(errHandler)

	err := closer.Close()
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCalls(c, nil)
}

func (s *multiCloserSuite) TestCloseError(c *gc.C) {
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
	closer := utils.NewMultiCloser(errHandler)
	closer.AddClosers(subCloserA, subCloserB, subCloserC)

	err := closer.Close()
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c, "close-A", "close-B", "errHandler", "close-C")
}

func (s *multiCloserSuite) TestCloseMultiError(c *gc.C) {
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
	closer := utils.NewMultiCloser(errHandler)
	closer.AddClosers(subCloserA, subCloserB, subCloserC)

	err := closer.Close()
	c.Assert(err, jc.ErrorIsNil)

	stub.CheckCallNames(c,
		"close-A", "errHandler",
		"close-B", "errHandler",
		"close-C", "errHandler",
	)
}
