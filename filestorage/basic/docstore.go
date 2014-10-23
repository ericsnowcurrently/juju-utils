// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package basic

import (
	"github.com/juju/errors"
	"github.com/juju/utils"

	"github.com/juju/utils/filestorage"
)

type docStorage struct {
	docs map[string]filestorage.Document
}

// NewDocStorage returns a simple memory-backed DocStorage.
func NewDocStorage() filestorage.DocStorage {
	storage := docStorage{
		docs: make(map[string]filestorage.Document),
	}
	return &storage
}

func (s *docStorage) lookUp(id string) (filestorage.Document, error) {
	doc, ok := s.docs[id]
	if !ok {
		return nil, errors.NotFoundf(id)
	}
	return doc, nil
}

// Doc implements DocStorage.Doc.
func (s *docStorage) Doc(id string) (filestorage.Document, error) {
	raw, err := s.lookUp(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return raw.Copy(id), nil
}

// ListDocs implements DocStorage.ListDocs.
func (s *docStorage) ListDocs() ([]filestorage.Document, error) {
	var list []filestorage.Document
	for _, doc := range s.docs {
		if doc == nil {
			continue
		}
		list = append(list, doc)
	}
	return list, nil
}

// AddDoc implements DocStorage.AddDoc.
func (s *docStorage) AddDoc(doc filestorage.Document) (string, error) {
	if doc.ID() != "" {
		return "", errors.AlreadyExistsf("ID already set")
	}

	uuid, err := utils.NewUUID()
	if err != nil {
		return "", errors.Annotate(err, "error while creating ID")
	}
	id := uuid.String()
	// We let the caller call meta.SetID() if they so desire.

	s.docs[id] = doc.Copy(id)
	return id, nil
}

// RemoveDoc implements DocStorage.RemoveDoc.
func (s *docStorage) RemoveDoc(id string) error {
	if _, ok := s.docs[id]; !ok {
		return errors.NotFoundf(id)
	}
	delete(s.docs, id)
	return nil
}

// Close implements io.Closer.Close.
func (s *docStorage) Close() error {
	return nil
}
