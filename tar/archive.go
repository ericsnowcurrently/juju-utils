// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var sep = string(os.PathSeparator)

//---------------------------
// Archive

//Archive identifies the files in a single tar archive.
type Archive struct {
	// Files identifies each file in the archive by relative path.
	Files []string

	// Root is the path to the root directory of the archive.
	Root string
}

// Write writes out the archive data for the files/directory-trees.
func (a *Archive) Write(w io.Writer) (err error) {
	checkClose := func(w io.Closer) {
		if closeErr := w.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing archive file: %v", closeErr)
		}
	}

	tarw := tar.NewWriter(w)
	defer checkClose(tarw)

	for _, ent := range a.Files {
		if err := a.writeTree(ent, tarw); err != nil {
			return fmt.Errorf("archive failed: %v", err)
		}
	}

	return nil
}

func (a *Archive) WriteGzipped(w io.Writer) (err error) {
	checkClose := func(w io.Closer) {
		if closeErr := w.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing archive file: %v", closeErr)
		}
	}

	gzw := gzip.NewWriter(w)
	defer checkClose(gzw)

	return a.Write(gzw)
}

func (a *Archive) Create(filename string) (err error) {
	checkClose := func(w io.Closer) {
		if closeErr := w.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing archive file: %v", closeErr)
		}
	}

	// Create the file.
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create archive file %q", filename)
	}
	defer checkClose(f)

	// Write out the archive.
	return a.Write(f)
}

// writeTree creates an entry for the given file
// or directory in the given tar archive.
func (a *Archive) writeTree(fileName string, tarw *tar.Writer) error {

	// Open and inspect the file.
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
}

//---------------------------
// Archiver

type Archiver struct {
	tarw *tar.Writer
}

func NewArchiver(tarfile io.Writer) *Archiver {
	ar := Archiver{
		tarw: tar.NewWriter(tarfile),
	}
	return &ar
}

// Add creates an entry for the given file/dir in the archive.
func (a *Archiver) Add(path, root string) (string, error) {
	name, info, file, err := a.open(path, root)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		name, err = a.addDirRaw(name, info, file, path)
	} else {
		name, err = a.AddFileRaw(name, info, file)
	}
	if err != nil {
		return "", err
	}
	return name, nil
}

// WriteFile creates an entry for the given file in the archive.
func (a *Archiver) AddTree(dirname, root string) error {
}

// AddFileRaw creates an entry for the given "file" in the archive.
func (a *Archiver) AddFileRaw(name string, info os.FileInfo, data io.Reader) (string, error) {
	hdr, err := a.writeHeader(name, info)
	if err != nil {
		return "", err
	}

	_, err := io.Copy(a.tarw, data)
	if err != nil {
		return "", fmt.Errorf("failed to write %q: %v", filename, err)
	}

	return hdr.Name, nil
}

func (a *Archiver) open(path, root string) (string, *os.File, *os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	name := strings.TrimPrefix(path, root)
	return name, info, file, err
}

func (a *Archiver) writeHeader(filename string, info os.FileInfo) (*tar.Header, error) {
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return "", fmt.Errorf("could not create tar header for %q: %v", filename, err)
	}
	hdr.Name = filepath.ToSlash(filename)

	// Write out the header.
	if err := a.tarw.WriteHeader(hdr); err != nil {
		return "", fmt.Errorf("could not write header for %q: %v", filename, err)
	}

	return hdr, nil
}

// writeTree creates an entry for the given file
// or directory in the given tar archive.
func (a *Archiver) WriteTree(fileName string, tarw *tar.Writer) error {
	fInfo, err := f.Stat()
	if err != nil {
		return err
	}
	h, err := tar.FileInfoHeader(fInfo, "")
	if err != nil {
		return fmt.Errorf("cannot create tar header for %q: %v", fileName, err)
	}
	h.Name = filepath.ToSlash(strings.TrimPrefix(fileName, a.StripPrefix))

	// Write out the header.
	if err := tarw.WriteHeader(h); err != nil {
		return fmt.Errorf("cannot write header for %q: %v", fileName, err)
	}

	// Write out the contents.
	if fInfo.IsDir() {
		return a.writeDir(fileName, f, tarw)
	} else {
		_, err := io.Copy(tarw, f)
		if err != nil {
			return fmt.Errorf("failed to write %q: %v", fileName, err)
		}
	}
	return nil
}

func (a *Archive) writeDir(dirname string, f *os.File, tarw *tar.Writer) error {
	if !strings.HasSuffix(dirname, sep) {
		dirname += sep
	}
	for {
		names, err := f.Readdirnames(100)
		if len(names) == 0 && err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading directory %q: %v", dirname, err)
		}
		for _, basename := range names {
			err := a.writeTree(filepath.Join(dirname, basename), tarw)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//---------------------------
// UnArchiver

type UnArchiver struct {
	tarr *tar.Reader
}

//...
