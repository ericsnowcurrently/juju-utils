// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package utils_test

import (
	"os"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils"
)

var _ = gc.Suite(&SignalSuite{})

type SignalSuite struct {
	testing.IsolationSuite

	stub     *testing.Stub
	signaler *stubSignaler
}

func (s *SignalSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.stub = &testing.Stub{}
	s.signaler = &stubSignaler{stub: s.stub}
}

func (s *SignalSuite) TestKillIfSupportedOkay(c *gc.C) {
	err := utils.KillIfSupported(s.signaler)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Kill")
}

func (s *SignalSuite) TestKillIfSupportedNotSupported(c *gc.C) {
	err := utils.KillIfSupported(s)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c)
}

func (s *SignalSuite) TestKillIfSupportedError(c *gc.C) {
	failure := errors.New("<failure>")
	s.stub.SetErrors(failure)

	err := utils.KillIfSupported(s.signaler)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.stub.CheckCallNames(c, "Kill")
}

func (s *SignalSuite) TestSignalInterrupt(c *gc.C) {
	supported, err := utils.Signal(s.signaler, os.Interrupt)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(supported, jc.IsTrue)
	s.stub.CheckCallNames(c, "Interrupt")
}

func (s *SignalSuite) TestSignalKill(c *gc.C) {
	supported, err := utils.Signal(s.signaler, os.Kill)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(supported, jc.IsTrue)
	s.stub.CheckCallNames(c, "Kill")
}

func (s *SignalSuite) TestSignalError(c *gc.C) {
	failure := errors.New("<failure>")
	s.stub.SetErrors(failure)

	supported, err := utils.Signal(s.signaler, os.Kill)

	c.Check(errors.Cause(err), gc.Equals, failure)
	c.Check(supported, jc.IsTrue)
	s.stub.CheckCallNames(c, "Kill")
}

func (s *SignalSuite) TestSignalNotSupported(c *gc.C) {
	supported, err := utils.Signal(s, os.Interrupt)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(supported, jc.IsFalse)
	s.stub.CheckNoCalls(c)
}

func (s *SignalSuite) TestSignalBadSignal(c *gc.C) {
	sig := &stubSignal{stub: s.stub}

	supported, err := utils.Signal(s.signaler, sig)

	c.Check(err, jc.Satisfies, errors.IsNotSupported)
	c.Check(supported, jc.IsFalse)
	s.stub.CheckCallNames(c, "String")
}

type stubSignaler struct {
	stub *testing.Stub
}

func (s *stubSignaler) Interrupt() error {
	s.stub.AddCall("Interrupt")
	if err := s.stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *stubSignaler) Kill() error {
	s.stub.AddCall("Kill")
	if err := s.stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type stubSignal struct {
	stub *testing.Stub

	ReturnString string
}

func (s stubSignal) String() string {
	s.stub.AddCall("String")
	s.stub.PopNoErr()

	return s.ReturnString
}

func (s stubSignal) Signal() {
	s.stub.AddCall("Signal")
	s.stub.PopNoErr()
}
