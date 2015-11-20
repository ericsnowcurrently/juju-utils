// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
)

var _ = gc.Suite(&ContextSuite{})

type ContextSuite struct {
	BaseSuite
}

func (s *ContextSuite) TestValidateOkay(c *gc.C) {
	ctx := exec.Context{
		Env: []string{"X=y"},
		Dir: "/x/y/z",
	}
	err := ctx.Validate()

	c.Check(err, jc.ErrorIsNil)
}

func (s *ContextSuite) TestValidateZeroValue(c *gc.C) {
	var ctx exec.Context
	err := ctx.Validate()

	c.Check(err, jc.ErrorIsNil)
}
