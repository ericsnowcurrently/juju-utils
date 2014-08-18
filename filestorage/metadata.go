// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package filestorage

import (
	"time"

	"github.com/juju/errors"
)

type Metadata interface {
	// ID is the unique ID assigned by the storage system.
	ID() string
	// Size is the size of the file (in bytes).
	Size() int64
	// Checksum is the checksum for the file.
	Checksum() string
	// ChecksumFormat is the kind (and encoding) of checksum.
	ChecksumFormat() string
	// Timestamp records when the file was created.
	Timestamp() time.Time
	// Stored indicates whether or not the file has been stored.
	Stored() bool

	// CheckComplete determines whether or not the metadata is complete.
	CheckComplete() error
	// Doc returns a storable copy of the metadata.
	Doc() interface{}

	// SetID sets the ID of the metadata.  If the ID is already set,
	// SetID() should return true (false otherwise).
	SetID(id string) (alreadySet bool)
	// SetFile sets the file info on the metadata.
	SetFile(size int64, checksum, checksumFormat string) error
	// SetStored sets Stored to true on the metadata.
	SetStored()
}

// Ensure FileMetadata implements Metadata.
var _ = Metadata((*FileMetadata)(nil))

// FileMetadata contains the metadata for a single stored file.
type FileMetadata struct {
	id        string
	timestamp time.Time
	finished  *FinishedMetadata
	stored    bool
}

// NewMetadata returns a new Metadata for a file.  ID is left unset (use
// SetID() for that).  Size, Checksum, and ChecksumFormat are left unset
// (use SetFile() for those).  If no timestamp is provided, the
// current one is used.
func NewMetadata(timestamp *time.Time) Metadata {
	meta := FileMetadata{}
	if timestamp == nil {
		meta.timestamp = time.Now().UTC()
	} else {
		meta.timestamp = *timestamp
	}
	return &meta
}

func (m *FileMetadata) ID() string {
	return m.id
}

func (m *FileMetadata) Size() int64 {
	if m.finished == nil {
		return 0
	}
	return m.finished.size
}

func (m *FileMetadata) Checksum() string {
	if m.finished == nil {
		return ""
	}
	return m.finished.checksum
}

func (m *FileMetadata) ChecksumFormat() string {
	if m.finished == nil {
		return ""
	}
	return m.finished.checksumFormat
}

func (m *FileMetadata) Timestamp() time.Time {
	return m.timestamp
}

func (m *FileMetadata) Stored() bool {
	return m.stored
}

func (m *FileMetadata) CheckComplete() error {
	if m.id == "" {
		return errors.New("missing ID")
	}
	if m.finished == nil {
		return errors.New("missing file info (see SetFile())")
	}
	return nil
}

func (m *FileMetadata) Doc() interface{} {
	return m
}

func (m *FileMetadata) SetID(id string) bool {
	if m.id != "" {
		return true
	}
	m.id = id
	return false
}

// FinishedMetadata contains the metadata specific to files that
// already exist.
type FinishedMetadata struct {
	size           int64
	checksum       string
	checksumFormat string
}

func (m *FileMetadata) SetFile(size int64, checksum, format string) error {
	if m.finished != nil {
		return errors.New("metadata already complete")
	}

	if size <= 0 {
		return errors.New("missing size")
	}
	if checksum == "" {
		return errors.New("missing checksum")
	}
	if format == "" {
		return errors.New("missing checksum format")
	}

	finished := FinishedMetadata{
		size:           size,
		checksum:       checksum,
		checksumFormat: format,
	}
	m.finished = &finished
	return nil
}

func (m *FileMetadata) SetStored() {
	m.stored = true
}
