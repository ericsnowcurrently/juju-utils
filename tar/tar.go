// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// This package provides convenience helpers on top of archive/tar
// to be able to tar/untar files with a functionality closer
// to gnu tar command.
package tar

import (
	"archive/tar"
	"fmt"
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
	proxy := hash.NewSHA1Proxy(target)
	ar := Archiver{fileList, strip}

	err = ar.Write(proxy)
	if err != nil {
		return "", err
	}

	return proxy.Base64Sum(), nil
}

// TODO(ericsnow) Move createAndFill down (below UntarFiles).

func createAndFill(filePath string, mode int64, content io.Reader) error {
	fh, err := os.Create(filePath)
	defer fh.Close()
	if err != nil {
		return fmt.Errorf("some of the tar contents cannot be written to disk: %v", err)
	}
	_, err = io.Copy(fh, content)
	if err != nil {
		return fmt.Errorf("failed while reading tar contents: %v", err)
	}
	err = os.Chmod(fh.Name(), os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("cannot set proper mode on file %q: %v", filePath, err)
	}
	return nil
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
			return fmt.Errorf("failed while reading tar header: %v", err)
		}
		fullPath := filepath.Join(outputFolder, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(fullPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("cannot extract directory %q: %v", fullPath, err)
			}
		case tar.TypeSymlink:
			if err = symlink.New(hdr.Linkname, fullPath); err != nil {
				return fmt.Errorf("cannot extract symlink %q to %q: %v", hdr.Linkname, fullPath, err)
			}
			continue
		case tar.TypeReg, tar.TypeRegA:
			if err = createAndFill(fullPath, hdr.Mode, tr); err != nil {
				return fmt.Errorf("cannot extract file %q: %v", fullPath, err)
			}
		}
	}
	return nil
}
