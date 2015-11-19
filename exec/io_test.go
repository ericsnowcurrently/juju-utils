// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"bytes"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&IOSuite{})

type IOSuite struct {
	BaseSuite
}

func (s *IOSuite) TestStderrToStdoutOkay(c *gc.C) {
	var stdout bytes.Buffer
	cmd := s.NewStubCommand()
	cmd.ReturnInfo = exec.CommandInfo{
		Context: exec.Context{
			Stdio: exec.Stdio{
				Out: &stdout,
			},
		},
	}

	err := exec.StderrToStdout(cmd)
	c.Assert(err, jc.ErrorIsNil)

	s.Stub.CheckCallNames(c, "Info", "SetStdio")
	s.Stub.CheckCall(c, 1, "SetStdio", exec.Stdio{
		Err: &stdout,
	})
}

func (s *IOSuite) TestStderrToStdoutNil(c *gc.C) {
	cmd := s.NewStubCommand()
	cmd.ReturnInfo = exec.CommandInfo{
		Context: exec.Context{
			Stdio: exec.Stdio{
				Out: nil,
			},
		},
	}

	err := exec.StderrToStdout(cmd)
	c.Assert(err, jc.ErrorIsNil)

	s.Stub.CheckCallNames(c, "Info", "SetStdio")
	s.Stub.CheckCall(c, 1, "SetStdio", exec.Stdio{})
}

func (s *IOSuite) TestStderrToStdoutError(c *gc.C) {
	failure := s.SetFailure()
	s.Stub.SetErrors(nil, failure)
	cmd := s.NewStubCommand()

	err := exec.StderrToStdout(cmd)

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Info", "SetStdio")
}
