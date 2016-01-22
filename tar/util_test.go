// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar_test

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type expectedTarContents []expectedTarContent

func (ec expectedTarContents) check(c *gc.C, rootDir string, contents map[string]string) {
	for i, expectedContent := range ec {
		c.Logf("(%d) trying %q", i, expectedContent.Name)

		name, ok := expectedContent.checkExists(c, rootDir)
		if !ok {
			continue
		}

		content, ok := contents[name]
		c.Logf("checking for presence of %q on untar files", name)
		if !c.Check(ok, jc.IsTrue) {
			continue
		}

		expectedContent.checkContent(c, content)
	}
}

type expectedTarContent struct {
	Name string
	Body string
}

func (ec expectedTarContent) checkExists(c *gc.C, rootDir string) (string, bool) {
	name := strings.TrimPrefix(ec.Name, string(os.PathSeparator))
	if rootDir != "" {
		expectedPath := filepath.Join(rootDir, name)
		_, err := os.Lstat(expectedPath)
		if !c.Check(err, gc.Equals, nil) {
			return "", false
		}
	}
	return name, true
}

func (ec expectedTarContent) checkContent(c *gc.C, content string) {
	if ec.Body == "" {
		return
	}
	c.Log("Also checking the file contents")
	c.Check(content, gc.Equals, ec.Body)
}

type testDirWalker struct {
	rootDir  string
	log      func(...interface{})
	contents map[string]string
}

func newTestDirWalker(c *gc.C, rootDir string) *testDirWalker {
	return &testDirWalker{
		rootDir: rootDir,
		log:     c.Log,
	}
}

func (dw *testDirWalker) readTar(c *gc.C, reader io.Reader) {
	c.Assert(dw.contents, gc.IsNil)
	dw.contents = make(map[string]string)

	tr := tar.NewReader(reader)
	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		c.Assert(err, gc.IsNil)

		data, err := ioutil.ReadAll(tr)
		c.Assert(err, gc.IsNil)

		dw.contents[hdr.Name] = string(data)
	}
}

func (dw *testDirWalker) walk(c *gc.C) {
	c.Assert(dw.contents, gc.IsNil)
	dw.contents = make(map[string]string)

	err := filepath.Walk(dw.rootDir, dw.handleFile)
	c.Assert(err, jc.ErrorIsNil)
}

func (dw *testDirWalker) handleFile(path string, finfo os.FileInfo, err error) error {
	fileName := strings.TrimPrefix(path, dw.rootDir)
	fileName = strings.TrimPrefix(fileName, string(os.PathSeparator))
	dw.log(fileName)

	if err != nil {
		return err
	}

	if fileName == "" {
		return nil
	}

	if finfo.Mode()&os.ModeType != 0 {
		dw.contents[fileName] = ""
		return nil
	}

	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	dw.contents[fileName] = string(data)

	return nil
}
