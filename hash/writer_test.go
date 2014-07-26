// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package hash_test

import (
	"bytes"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/filetesting"
	gc "gopkg.in/check.v1"

	"github.com/juju/utils/hash"
)

var _ = gc.Suite(&WriterSuite{})

type WriterSuite struct {
	testing.IsolationSuite

	stub    *testing.Stub
	wBuffer *bytes.Buffer
	writer  *filetesting.StubWriter
	hBuffer *bytes.Buffer
	hash    *filetesting.StubHash
}

func (s *WriterSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)

	s.stub = &testing.Stub{}
	s.wBuffer = new(bytes.Buffer)
	s.writer = &filetesting.StubWriter{
		Stub:        s.stub,
		ReturnWrite: s.wBuffer,
	}
	s.hBuffer = new(bytes.Buffer)
	s.hash = filetesting.NewStubHash(s.stub, s.hBuffer)
}

func (s *WriterSuite) TestHashingWriterWriteEmpty(c *gc.C) {
	w := hash.NewHashingWriter(s.writer, s.hash)
	n, err := w.Write(nil)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Write", "Write")
	c.Check(n, gc.Equals, 0)
	c.Check(s.wBuffer.String(), gc.Equals, "")
	c.Check(s.hBuffer.String(), gc.Equals, "")
}

func (s *WriterSuite) TestHashingWriterWriteSmall(c *gc.C) {
	w := hash.NewHashingWriter(s.writer, s.hash)
	n, err := w.Write([]byte("spam"))
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Write", "Write")
	c.Check(n, gc.Equals, 4)
	c.Check(s.wBuffer.String(), gc.Equals, "spam")
	c.Check(s.hBuffer.String(), gc.Equals, "spam")
}

func (s *WriterSuite) TestHashingWriterWriteFileError(c *gc.C) {
	w := hash.NewHashingWriter(s.writer, s.hash)
	failure := errors.New("<failed>")
	s.stub.SetErrors(failure)

	_, err := w.Write([]byte("spam"))

	s.stub.CheckCallNames(c, "Write")
	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *WriterSuite) TestHashingWriterWriteHasherError(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{}
	hasher.err = errors.New("failed!")
	w := hash.NewHashingWriter(&file, &hasher)
	_, err := w.Write([]byte("spam"))

	c.Check(err, gc.ErrorMatches, "failed!")
}

func (s *WriterSuite) TestHashingWriterBase64Sum(c *gc.C) {
	s.hash.ReturnSum = []byte("spam")
	w := hash.NewHashingWriter(s.writer, s.hash)
	b64sum := w.Base64Sum()

	s.stub.CheckCallNames(c, "Sum")
	c.Check(b64sum, gc.Equals, "c3BhbQ==")
}

func (s *WriterSuite) TestHashingWriterHash(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&file, &hasher)
	b64hash := w.Hash()

	c.Check(b64hash, gc.Equals, "c3BhbQ==")
}

func (s *WriterSuite) TestHashingWriterRawHash(c *gc.C) {
	file := fakeFile{}
	hasher := fakeHasher{sum: []byte("spam")}
	w := hash.NewHashingWriter(&file, &hasher)
	rawhash := w.RawHash()

	c.Check(rawhash, gc.Equals, "7370616d")
}

type fakeFile struct {
	written []byte
	err     error
}

func (f *fakeFile) Write(data []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.written == nil {
		f.written = make([]byte, 0)
	}
	f.written = append(f.written, data...)
	return len(data), nil
}

type fakeHasher struct {
	fakeFile
	sum []byte
}

func (h *fakeHasher) Sum(b []byte) []byte {
	return h.sum
}

// Not used:
func (h *fakeHasher) Reset()         {}
func (h *fakeHasher) Size() int      { return -1 }
func (h *fakeHasher) BlockSize() int { return -1 }
