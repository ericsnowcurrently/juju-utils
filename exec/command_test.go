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

var (
	_ = gc.Suite(&CommandSuite{})
	_ = gc.Suite(&CommandInfoSuite{})
)

type CommandSuite struct {
	BaseSuite
}

func (s *CommandSuite) TestNewCommandOkay(c *gc.C) {
	expected := s.NewStubCommand()
	s.StubExec.ReturnCommand = expected

	cmd, err := exec.NewCommand(s.StubExec, "/x/y/z/spam", "--ham", "eggs")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(cmd, gc.Equals, expected)
	s.Stub.CheckCallNames(c, "Command")
	s.Stub.CheckCall(c, 0, "Command",
		exec.CommandInfo{
			Path: "/x/y/z/spam",
			Args: []string{
				"/x/y/z/spam",
				"--ham",
				"eggs",
			},
		},
	)
}

func (s *CommandSuite) TestNewCommandError(c *gc.C) {
	expected := s.NewStubCommand()
	s.StubExec.ReturnCommand = expected
	failure := s.SetFailure()

	_, err := exec.NewCommand(s.StubExec, "/x/y/z/spam", "--ham", "eggs")

	c.Check(errors.Cause(err), gc.Equals, failure)
	s.Stub.CheckCallNames(c, "Command")
}

type CommandInfoSuite struct {
	BaseSuite
}

func (s *CommandInfoSuite) TestNewCommandInfo(c *gc.C) {
	info := exec.NewCommandInfo("/x/y/z/spam", "--ham", "eggs")

	c.Check(info, jc.DeepEquals, exec.CommandInfo{
		Path: "/x/y/z/spam",
		Args: []string{
			"/x/y/z/spam",
			"--ham",
			"eggs",
		},
	})
}

func (s *CommandInfoSuite) TestStringOkay(c *gc.C) {
	info := exec.NewCommandInfo("/x/y/z/spam", "--ham", "eggs")
	str := info.String()

	c.Check(str, gc.Equals, "/x/y/z/spam --ham eggs")
}

func (s *CommandInfoSuite) TestStringFull(c *gc.C) {
	info := exec.CommandInfo{
		Path: "/x/y/z/spam",
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
		Context: exec.Context{
			Env: []string{"X=y"},
			Dir: "/x/y/z",
			Stdio: exec.Stdio{
				In:  &bytes.Buffer{},
				Out: &bytes.Buffer{},
				Err: &bytes.Buffer{},
			},
		},
	}
	str := info.String()

	c.Check(str, gc.Equals, "/x/y/z/spam --ham eggs")
}

func (s *CommandInfoSuite) TestStringDifferentPath(c *gc.C) {
	info := exec.CommandInfo{
		Path: "/x/y/z/spam",
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
	}
	str := info.String()

	c.Check(str, gc.Equals, "/x/y/z/spam --ham eggs")
}

func (s *CommandInfoSuite) TestStringNoPath(c *gc.C) {
	info := exec.CommandInfo{
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
	}
	str := info.String()

	c.Check(str, gc.Equals, "spam --ham eggs")
}

func (s *CommandInfoSuite) TestStringNoArgs(c *gc.C) {
	info := exec.CommandInfo{
		Path: "/x/y/z/spam",
	}
	str := info.String()

	c.Check(str, gc.Equals, "/x/y/z/spam")
}

func (s *CommandInfoSuite) TestValidateOkay(c *gc.C) {
	renderer := s.newStubRenderer()
	renderer.ReturnBase = "/x/y/z/spam"
	info := exec.NewCommandInfo("/x/y/z/spam", "--ham", "eggs")

	err := info.ValidateRendered(renderer)

	c.Check(err, jc.ErrorIsNil)
}

func (s *CommandInfoSuite) TestValidateFull(c *gc.C) {
	renderer := s.newStubRenderer()
	renderer.ReturnBase = "spam"
	info := exec.CommandInfo{
		Path: "/x/y/z/spam",
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
		Context: exec.Context{
			Env: []string{"X=y"},
			Dir: "/x/y/z",
			Stdio: exec.Stdio{
				In:  &bytes.Buffer{},
				Out: &bytes.Buffer{},
				Err: &bytes.Buffer{},
			},
		},
	}

	err := info.ValidateRendered(renderer)

	c.Check(err, jc.ErrorIsNil)
}

func (s *CommandInfoSuite) TestValidateMissingPath(c *gc.C) {
	info := exec.CommandInfo{
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
	}

	err := info.ValidateRendered(nil)

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *CommandInfoSuite) TestValidateMissingArgs(c *gc.C) {
	info := exec.CommandInfo{
		Path: "/x/y/z/spam",
	}

	err := info.ValidateRendered(nil)

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *CommandInfoSuite) TestValidateCommandNameMismatch(c *gc.C) {
	renderer := s.newStubRenderer()
	renderer.ReturnBase = "foo"
	info := exec.CommandInfo{
		Path: "/x/y/z/foo",
		Args: []string{
			"spam",
			"--ham",
			"eggs",
		},
	}

	err := info.ValidateRendered(renderer)

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *CommandInfoSuite) TestValidateBadContext(c *gc.C) {
	c.Skip("for now there are no bad contexts")
}
