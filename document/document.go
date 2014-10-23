// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package document

import (
	"encoding/json"
	"io"
	"time"

	"github.com/juju/errors"
)

// RawDoc is a basic, uniquely identifiable document.
type RawDoc struct {
	// ID is the unique identifier for the document.
	ID string
	// Created is when the document was created.
	Created time.Time
}

// Doc is the obvious implementation of Document.  While perhaps useful
// on its own, it is most useful for embedding in other types.
type Doc struct {
	// Raw holds the raw data backing doc.
	Raw *RawDoc
}

// NewDocument returns a new Document.  ID is left unset (use SetID()
// for that).  If no created is provided, the current one is used.
func NewDocument(created *time.Time) *Doc {
	doc := Doc{
		Raw: &RawDoc{},
	}
	if created == nil {
		doc.Raw.Created = time.Now().UTC()
	} else {
		doc.Raw.Created = *created
	}
	return &doc
}

// ID returns the document's unique identifier.
func (d *Doc) ID() string {
	return d.Raw.ID
}

// Created implements Doc.Created.
func (d *Doc) Created() time.Time {
	return d.Raw.Created
}

// SetID sets the document's unique identifier.  If the ID is already
// set, SetID() returns true (false otherwise).
func (d *Doc) SetID(id string) bool {
	if d.Raw.ID != "" {
		return true
	}
	d.Raw.ID = id
	return false
}

// Copy returns a new Doc with Raw set to a shallow copy of the current
// value.  The raw ID is set to the one passed in.
func (d *Doc) Copy(id string) Document {
	raw := *d.Raw
	raw.ID = id
	return &Doc{&raw}
}

// Dump writes out the serialized doc.  Only the "json" format is
// supported.
func (m *Doc) Dump(w io.Writer, format string) error {
	switch format {
	case "json", "JSON":
		return errors.Trace(json.NewEncoder(w).Encode(m))
	default:
		return errors.NotSupportedf("format: %q", format)
	}
}

// Load updates the doc from serialized data in the reader.  Only the
// "json" format is supported.
func (m *Doc) Load(r io.Reader, format string) error {
	switch format {
	case "json", "JSON":
		return errors.Trace(json.NewDecoder(r).Decode(m))
	default:
		return errors.NotSupportedf("format: %q", format)
	}
}

// DefaultID returns an ID string derived from the doc.
func (m *Doc) DefaultID() (string, error) {
	return "", errors.NotSupportedf("no default ID")
}

// Validate checks the doc at the specified level.
func (m *Doc) Validate(level string) error {
	if level == "" {
		level = "full"
	}

	// Each case (except nop) should fall through to the next one.  Thus
	// the more restrictive the level, the higher up it should be in the
	// switch.
	switch level {
	case "full":
		fallthrough
	case "id":
		if m.Raw.ID == "" {
			return errors.NotValidf("missing ID")
		}
	case "initialized":
		if m.Raw.Created.IsZero() {
			return errors.NotValidf("missing Created")
		}
		// Don't fall through to the default.
	default:
		return errors.NotSupportedf("unrecognized level: %q", level)
	}

	return nil
}
