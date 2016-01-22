// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	tarutil "github.com/juju/utils/tar"
)

var _ = gc.Suite(&TarSuite{})

type TarSuite struct {
	testing.IsolationSuite
	cwd       string
	testFiles []string
}

func (t *TarSuite) SetUpTest(c *gc.C) {
	t.cwd = c.MkDir()
	t.IsolationSuite.SetUpTest(c)
}

func (t *TarSuite) TearDownTest(c *gc.C) {
	t.testFiles = nil

	t.IsolationSuite.TearDownTest(c)
}

func (t *TarSuite) createTestFiles(c *gc.C) {
	c.Assert(t.testFiles, gc.HasLen, 0)
	t.testFiles = testExpectedTarContents.create(c, t.cwd)
}

func (t *TarSuite) removeTestFiles(c *gc.C) {
	for _, removable := range t.testFiles {
		err := os.RemoveAll(removable)
		c.Check(err, jc.ErrorIsNil)
	}
	t.testFiles = nil
}

func (t *TarSuite) TestTarFiles(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)

	shaSum, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, jc.ErrorIsNil)

	outputBytes := outputTar.Bytes()
	fileShaSum := shaSumFile(c, bytes.NewBuffer(outputBytes))
	c.Check(shaSum, gc.Equals, fileShaSum)
	dw := newTestDirWalker(c, "")
	dw.readTar(c, bytes.NewBuffer(outputBytes))
	testExpectedTarContents.check(c, "", dw.contents)
}

func (t *TarSuite) TestSymlinksTar(c *gc.C) {
	expectedContents := expectedTarContents{
		{"TarDirectory", ""},
		{"TarLink", ""},
	}
	testFiles := expectedContents.create(c, t.cwd)
	tarDirP := testFiles[0]

	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)

	_, err := tarutil.TarFiles(testFiles, &outputTar, trimPath)
	c.Assert(err, jc.ErrorIsNil)

	outputBytes := outputTar.Bytes()
	tr := tar.NewReader(bytes.NewBuffer(outputBytes))
	symlinks := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		c.Assert(err, jc.ErrorIsNil)
		if hdr.Typeflag == tar.TypeSymlink {
			symlinks += 1
			c.Check(hdr.Linkname, gc.Equals, tarDirP)
		}
	}
	c.Check(symlinks, gc.Equals, 1)
}

func (t *TarSuite) TestUnTarFilesUncompressed(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, jc.ErrorIsNil)
	t.removeTestFiles(c) // ...not strictly necessary.

	outputDir := filepath.Join(t.cwd, "TarOuputFolder")
	err = os.Mkdir(outputDir, 0755)
	c.Assert(err, jc.ErrorIsNil)

	err = tarutil.UntarFiles(&outputTar, outputDir)
	c.Assert(err, jc.ErrorIsNil)

	dw := newTestDirWalker(c, outputDir)
	dw.walk(c)
	testExpectedTarContents.check(c, outputDir, dw.contents)
}

func (t *TarSuite) TestFindFileFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, jc.ErrorIsNil)
	t.removeTestFiles(c) // ...not strictly necessary.

	_, file, err := tarutil.FindFile(&outputTar, "TarDirectoryPopulated/TarSubFile1")
	c.Assert(err, jc.ErrorIsNil)

	data, err := ioutil.ReadAll(file)
	c.Assert(err, jc.ErrorIsNil)
	c.Check(string(data), gc.Equals, "TarSubFile1")
}

func (t *TarSuite) TestFindFileNotFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, jc.ErrorIsNil)
	t.removeTestFiles(c) // ...not strictly necessary.

	_, _, err = tarutil.FindFile(&outputTar, "does_not_exist")

	c.Check(err, gc.ErrorMatches, "does_not_exist not found")
}

func (t *TarSuite) TestUntarFilesHeadersIgnored(c *gc.C) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	err := w.WriteHeader(&tar.Header{
		Name:     "pax_global_header",
		Typeflag: tar.TypeXGlobalHeader,
	})
	c.Assert(err, jc.ErrorIsNil)
	err = w.Flush()
	c.Assert(err, jc.ErrorIsNil)

	err = tarutil.UntarFiles(&buf, t.cwd)
	c.Assert(err, jc.ErrorIsNil)

	err = filepath.Walk(t.cwd, func(path string, finfo os.FileInfo, err error) error {
		if path != t.cwd {
			return fmt.Errorf("unexpected file: %v", path)
		}
		return err
	})
	c.Assert(err, jc.ErrorIsNil)
}

var testExpectedTarContents = expectedTarContents{
	{"TarDirectoryEmpty", ""},
	{"TarDirectoryPopulated", ""},
	{"TarLink", ""},
	{"TarDirectoryPopulated/TarSubFile1", "TarSubFile1"},
	{"TarDirectoryPopulated/TarSubLink", ""},
	{"TarDirectoryPopulated/TarDirectoryPopulatedSubDirectory", ""},
	{"TarFile1", "TarFile1"},
	{"TarFile2", "TarFile2"},
}
