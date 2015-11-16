// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&runSuite{})

type runSuite struct {
	BaseSuite
}

func (s *runSuite) TestRunOkay(c *gc.C) {
	expected := s.NewStubProcessState()
	cmd := s.NewStubCommand()
	process := s.NewStubProcess()
	process.ReturnWait = expected
	cmd.ReturnStart = process
	state, err := exec.Run(cmd)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(state, gc.Equals, expected)
	s.Stub.CheckCallNames(c, "Start", "Wait")
}

func (s *runSuite) TestRunErrorStart(c *gc.C) {
	failure := s.SetFailure()
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.Run(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Start")
}

func (s *runSuite) TestRunErrorWait(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, failure)
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.Run(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Start", "Wait")
}

func (s *runSuite) TestOutputOkay(c *gc.C) {
	var input string
	cmd := s.newStdioCommand(&input,
		"abc",
		"!xyz",
	)

	data, err := exec.Output(cmd)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(input, gc.Equals, "")
	c.Check(string(data), gc.Equals, "abc")
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}

func (s *runSuite) TestOutputErrorSetStdio(c *gc.C) {
	failure := s.SetFailure()
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.Output(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio")
}

func (s *runSuite) TestOutputErrorStart(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, failure)
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.Output(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio", "Start")
}

func (s *runSuite) TestOutputErrorWait(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, nil, failure)
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.Output(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}

func (s *runSuite) TestCombinedOutputOkay(c *gc.C) {
	var input string
	cmd := s.newStdioCommand(&input,
		"abc",
		"!xyz",
	)

	data, err := exec.CombinedOutput(cmd)
	c.Assert(err, jc.ErrorIsNil)

	c.Check(input, gc.Equals, "")
	c.Check(string(data), gc.Equals, "abcxyz")
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}

func (s *runSuite) TestCombinedOutputErrorSetStdio(c *gc.C) {
	failure := s.SetFailure()
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.CombinedOutput(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio")
}

func (s *runSuite) TestCombinedOutputErrorStart(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, failure)
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.CombinedOutput(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio", "Start")
}

func (s *runSuite) TestCombinedOutputErrorWait(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, nil, failure)
	cmd := s.NewStubCommand()
	cmd.ReturnStart = s.NewStubProcess()
	_, err := exec.CombinedOutput(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "SetStdio", "Start", "Wait")
}