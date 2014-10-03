// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package document

// RawDoc is a basic, uniquely identifiable document.
type RawDoc struct {
	// ID is the unique identifier for the document.
	ID string
}

// Doc is the obvious implementation of Document.  While perhaps useful
// on its own, it is most useful for embedding in other types.
type Doc struct {
	// Raw holds the raw data backing doc.
	Raw *RawDoc
}

// ID returns the document's unique identifier.
func (d *Doc) ID() string {
	return d.Raw.ID
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
	copied := *d.Raw
	copied.ID = id
	return &Doc{&copied}
}
