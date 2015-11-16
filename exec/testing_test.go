// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"io/ioutil"
	"strings"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
	"github.com/juju/utils/exec/exectesting"
)

type BaseSuite struct {
	testing.IsolationSuite
	exectesting.StubSuite
	exec.TestingExposer
}

func (s *BaseSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.StubSuite.SetUpTest(c)
}

func (s *BaseSuite) SetExecPIDs(e exec.Exec, pids ...int) {
	var processes []exec.Process
	for _, pid := range pids {
		process := s.NewStubProcess()
		process.ReturnPID = pid
		processes = append(processes, process)
	}
	s.SetExec(e, processes...)
}

func (s *BaseSuite) newStdioCommand(input *string, output ...string) exec.Command {

	return s.NewStdioCommand(func(stdio exec.Stdio, origErr error) error {
		// TODO(ericsnow) Conditionally handle origErr?

		data, err := ioutil.ReadAll(stdio.In)
		if err != nil {
			return err
		}
		*input = string(data)

		for _, out := range output {
			if strings.HasPrefix(out, "!") {
				if _, err := stdio.Err.Write([]byte(out[1:])); err != nil {
					return err
				}
			} else {
				if _, err := stdio.Out.Write([]byte(out)); err != nil {
					return err
				}
			}
		}

		return origErr
	})

}
