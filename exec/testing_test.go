// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"io/ioutil"
	"os"
	osfilepath "path/filepath"
	"strings"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/exec"
	"github.com/juju/utils/exec/exectesting"
	"github.com/juju/utils/filepath"
)

type BaseSuite struct {
	testing.IsolationSuite
	exectesting.StubSuite
}

func (s *BaseSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.StubSuite.SetUpTest(c)
}

func (s *BaseSuite) AddDir(c *gc.C, path ...string) string {
	root := c.MkDir()
	dirname := osfilepath.Join(append([]string{root}, path...)...)

	err := os.MkdirAll(dirname, 0755)
	c.Assert(err, jc.ErrorIsNil)

	return dirname
}

func (s *BaseSuite) AddBinDir(c *gc.C, path ...string) string {
	if len(path) == 0 {
		path = append(path, "bin")
	}

	dirname := s.AddDir(c, path...)

	s.PatchEnvPathPrepend(dirname)

	return dirname
}

func (s *BaseSuite) AddScript(c *gc.C, name, script string) string {
	binDir := s.AddBinDir(c)
	filename := osfilepath.Join(binDir, name)

	err := ioutil.WriteFile(filename, []byte(script), 0755)
	c.Assert(err, jc.ErrorIsNil)

	return filename
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

func (s *BaseSuite) newStubRenderer() *stubRenderer {
	return &stubRenderer{stub: s.Stub}
}

type stubRenderer struct {
	filepath.Renderer
	stub *testing.Stub

	ReturnBase string
}

func (s *stubRenderer) Base(path string) string {
	s.stub.AddCall("Base", path)
	s.stub.PopNoErr()

	return s.ReturnBase
}
