// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar_test

import (
	"archive/tar"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juju/testing"
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
	removeTestFiles(c, t.testFiles)

	t.IsolationSuite.TearDownTest(c)
}

func (t *TarSuite) createTestFiles(c *gc.C) {
	t.testFiles = createTestFiles(c, t.cwd)
}

func (t *TarSuite) removeTestFiles(c *gc.C) {
	removeTestFiles(c, t.testFiles)
	t.testFiles = nil
}

func shaSumFile(c *gc.C, fileToSum io.Reader) string {
	shahash := sha1.New()
	_, err := io.Copy(shahash, fileToSum)
	c.Assert(err, gc.IsNil)
	return base64.StdEncoding.EncodeToString(shahash.Sum(nil))
}

func (t *TarSuite) TestTarFiles(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	shaSum, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)
	outputBytes := outputTar.Bytes()
	fileShaSum := shaSumFile(c, bytes.NewBuffer(outputBytes))
	c.Assert(shaSum, gc.Equals, fileShaSum)
	checkTarContents(c, testExpectedTarContents, bytes.NewBuffer(outputBytes))
}

func (t *TarSuite) TestSymlinksTar(c *gc.C) {
	tarDirP := filepath.Join(t.cwd, "TarDirectory")
	err := os.Mkdir(tarDirP, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarlink1 := filepath.Join(t.cwd, "TarLink")
	err = os.Symlink(tarDirP, tarlink1)
	c.Check(err, gc.IsNil)
	testFiles := []string{tarDirP, tarlink1}

	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err = tarutil.TarFiles(testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)

	outputBytes := outputTar.Bytes()
	tr := tar.NewReader(bytes.NewBuffer(outputBytes))
	symlinks := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		c.Assert(err, gc.IsNil)
		if hdr.Typeflag == tar.TypeSymlink {
			symlinks += 1
			c.Assert(hdr.Linkname, gc.Equals, tarDirP)
		}
	}
	c.Assert(symlinks, gc.Equals, 1)

}

func (t *TarSuite) TestUnTarFilesUncompressed(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Check(err, gc.IsNil)
	t.removeTestFiles(c) // ...not strictly necessary.

	outputDir := filepath.Join(t.cwd, "TarOuputFolder")
	err = os.Mkdir(outputDir, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarutil.UntarFiles(&outputTar, outputDir)
	checkFilesWhereUntarred(c, testExpectedTarContents, outputDir)
}

func (t *TarSuite) TestFindFileFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, gc.IsNil)
	t.removeTestFiles(c) // ...not strictly necessary.

	_, file, err := tarutil.FindFile(&outputTar, "TarDirectoryPopulated/TarSubFile1")
	c.Assert(err, gc.IsNil)

	data, err := ioutil.ReadAll(file)
	c.Assert(err, gc.IsNil)

	c.Check(string(data), gc.Equals, "TarSubFile1")
}

func (t *TarSuite) TestFindFileNotFound(c *gc.C) {
	t.createTestFiles(c)
	var outputTar bytes.Buffer
	trimPath := fmt.Sprintf("%s/", t.cwd)
	_, err := tarutil.TarFiles(t.testFiles, &outputTar, trimPath)
	c.Assert(err, gc.IsNil)
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
	c.Assert(err, gc.IsNil)
	err = w.Flush()
	c.Assert(err, gc.IsNil)

	err = tarutil.UntarFiles(&buf, t.cwd)
	err = filepath.Walk(t.cwd, func(path string, finfo os.FileInfo, err error) error {
		if path != t.cwd {
			return fmt.Errorf("unexpected file: %v", path)
		}
		return err
	})
	c.Assert(err, gc.IsNil)
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

func createTestFiles(c *gc.C, cwd string) []string {
	tarDirE := filepath.Join(cwd, "TarDirectoryEmpty")
	err := os.Mkdir(tarDirE, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarDirP := filepath.Join(cwd, "TarDirectoryPopulated")
	err = os.Mkdir(tarDirP, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarlink1 := filepath.Join(cwd, "TarLink")
	err = os.Symlink(tarDirP, tarlink1)
	c.Check(err, gc.IsNil)

	tarSubFile1 := filepath.Join(tarDirP, "TarSubFile1")
	tarSubFile1Handle, err := os.Create(tarSubFile1)
	c.Check(err, gc.IsNil)
	tarSubFile1Handle.WriteString("TarSubFile1")
	tarSubFile1Handle.Close()

	tarSublink1 := filepath.Join(tarDirP, "TarSubLink")
	err = os.Symlink(tarSubFile1, tarSublink1)
	c.Check(err, gc.IsNil)

	tarSubDir := filepath.Join(tarDirP, "TarDirectoryPopulatedSubDirectory")
	err = os.Mkdir(tarSubDir, os.FileMode(0755))
	c.Check(err, gc.IsNil)

	tarFile1 := filepath.Join(cwd, "TarFile1")
	tarFile1Handle, err := os.Create(tarFile1)
	c.Check(err, gc.IsNil)
	tarFile1Handle.WriteString("TarFile1")
	tarFile1Handle.Close()

	tarFile2 := filepath.Join(cwd, "TarFile2")
	tarFile2Handle, err := os.Create(tarFile2)
	c.Check(err, gc.IsNil)
	tarFile2Handle.WriteString("TarFile2")
	tarFile2Handle.Close()

	return []string{
		tarDirE,
		tarDirP,
		tarlink1,
		tarFile1,
		tarFile2,
	}
}

func removeTestFiles(c *gc.C, testFiles []string) {
	for _, removable := range testFiles {
		err := os.RemoveAll(removable)
		c.Assert(err, gc.IsNil)
	}
}

// checkTarContents checks that the tar reader provided contains
// the expected files.
// expectedContents: a slice of the filenames with relative paths that are
// expected to be on the tar file.
// tarFile: the path of the file to be checked.
func checkTarContents(c *gc.C, expectedContents expectedTarContents, tarFile io.Reader) {
	dw := newTestDirWalker(c, "")
	dw.readTar(c, tarFile)

	expectedContents.check(c, "", dw.contents)
}

func checkFilesWhereUntarred(c *gc.C, expectedContents expectedTarContents, tarOutputFolder string) {
	dw := newTestDirWalker(c, tarOutputFolder)
	dw.walk(c)

	expectedContents.check(c, tarOutputFolder, dw.contents)
}
