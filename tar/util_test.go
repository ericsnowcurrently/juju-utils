// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package tar_test

import (
	"archive/tar"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func shaSumFile(c *gc.C, fileToSum io.Reader) string {
	shahash := sha1.New()
	_, err := io.Copy(shahash, fileToSum)
	c.Assert(err, jc.ErrorIsNil)
	return base64.StdEncoding.EncodeToString(shahash.Sum(nil))
}

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

func (ec expectedTarContents) create(c *gc.C, cwd string) []string {
	c.Logf("creating in %q", cwd)
	var topLevel []string
	for i, expectedContent := range ec {
		c.Logf("(%d) creating %q", i, expectedContent.Name)

		var source string
		if strings.Contains(expectedContent.Name, "Link") {
			source = filepath.Join(cwd, ec[i-1].Name)
			c.Logf("  symlink to %q", source)
		}
		path := expectedContent.create(c, cwd, source)
		if !strings.Contains(expectedContent.Name, "/") {
			topLevel = append(topLevel, path)
		}
	}
	return topLevel
}

type expectedTarContent struct {
	Name string
	Body string
}

func (ec expectedTarContent) create(c *gc.C, cwd, source string) string {
	path := filepath.Join(cwd, ec.Name)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	c.Assert(err, jc.ErrorIsNil)

	switch {
	case source != "":
		err := os.Symlink(source, path)
		c.Assert(err, jc.ErrorIsNil)
	case ec.Body != "":
		file, err := os.Create(path)
		c.Assert(err, jc.ErrorIsNil)
		defer file.Close()
		_, err = file.WriteString(ec.Body)
		c.Assert(err, jc.ErrorIsNil)
	default:
		err := os.MkdirAll(path, 0755)
		c.Assert(err, jc.ErrorIsNil)
	}
	return path
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
		c.Assert(err, jc.ErrorIsNil)

		data, err := ioutil.ReadAll(tr)
		c.Assert(err, jc.ErrorIsNil)

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
