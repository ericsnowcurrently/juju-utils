// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// This package provides convenience helpers on top of archive/tar
// to be able to tar/untar files with a functionality closer
// to gnu tar command.
package tar

import (
	"archive/tar"
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"

	"github.com/juju/errors"

	"github.com/juju/utils/hash"
	"github.com/juju/utils/symlink"
)

// FindFile returns the header and ReadCloser for the entry in the
// tarfile that matches the filename.  If nothing matches, an
// errors.NotFound error is returned.
func FindFile(tarFile io.Reader, filename string) (*tar.Header, io.Reader, error) {
	reader := tar.NewReader(tarFile)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		if header.Name == filename {
			return header, reader, nil
		}
	}

	return nil, nil, errors.NotFoundf(filename)
}

// TarFiles writes a tar stream into target holding the files listed
// in fileList. strip will be removed from the beginning of all the paths
// when stored (much like gnu tar -C option)
// Returns a Sha sum of the tar and nil if everything went well
// or empty sting and error in case of error.
// We use a base64 encoded sha1 hash, because this is the hash
// used by RFC 3230 Digest headers in http responses
func TarFiles(fileList []string, target io.Writer, strip string) (shaSum string, err error) {
	h := sha1.New()
	proxy := io.MultiWriter(target, h)
	ar := Archiver{
		Files:       fileList,
		StripPrefix: strip,
	}

	err = ar.Write(proxy)
	if err != nil {
		return "", err
	}

	fp := hash.NewValidFingerprint(h)
	return fp.Base64(), nil
}

// UntarFiles will extract the contents of tarFile using
// outputFolder as root
func UntarFiles(tarFile io.Reader, outputFolder string) error {
	tr := tar.NewReader(tarFile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			return nil
		}
		if err != nil {
			return errors.Annotate(err, "failed while reading tar header")
		}
		fullPath := filepath.Join(outputFolder, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(fullPath, os.FileMode(hdr.Mode)); err != nil {
				return errors.Annotatef(err, "cannot extract directory %q", fullPath)
			}
		case tar.TypeSymlink:
			if err = symlink.New(hdr.Linkname, fullPath); err != nil {
				return errors.Annotatef(err, "cannot extract symlink %q to %q", hdr.Linkname, fullPath)
			}
			continue
		case tar.TypeReg, tar.TypeRegA:
			if err = createAndFill(fullPath, hdr.Mode, tr); err != nil {
				return errors.Annotatef(err, "cannot extract file %q", fullPath)
			}
		}
	}
	return nil
}

func createAndFill(filePath string, mode int64, content io.Reader) error {
	fh, err := os.Create(filePath)
	defer fh.Close()
	if err != nil {
		return errors.Annotate(err, "some of the tar contents cannot be written to disk")
	}
	_, err = io.Copy(fh, content)
	if err != nil {
		return errors.Annotate(err, "failed while reading tar contents")
	}
	err = os.Chmod(fh.Name(), os.FileMode(mode))
	if err != nil {
		return errors.Annotatef(err, "cannot set proper mode on file %q", filePath)
	}
	return nil
}

// Iter is an iterator of the entries in a tar file.
type Iter struct {
	reader *tar.Reader
	done   bool
	err    error
}

// IterTarFile returns a new tarfile iterator wrapping tarfile.
func IterTarFile(tarfile io.Reader) *Iter {
	iter := Iter{
		reader: tar.NewReader(tarfile),
	}
	return &iter
}

func (it *Iter) Done() bool {
	return it.done
}

func (it *Iter) Err() error {
	return it.err
}

// Next populates header with the next value and returns true.  If there
// are no more values or there is an error, it returns false.  In the
// case that an error is encounterd, the error is stored on the struct.
func (it *Iter) Next(header *tar.Header, data io.Reader) bool {
	if it.done {
		return false
	}

	// Advance to the next entry.
	hdr, err := it.reader.Next()
	if err == io.EOF {
		// end of archive
		it.done = true
		return false
	}
	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	*entry = *hdr
	*data = *it.reader
	return true
}
