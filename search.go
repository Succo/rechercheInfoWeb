package main

import (
	"encoding/gob"
	"os"
)

// Search stores information relavant to parsed documents
type Search struct {
	// For each token we store the id of the first document where it was seen for heap law
	Token map[string]int
	Index map[string][]int
	Size  int
}

// NewSearch generates a Search loading a serialized file
func NewSearch(filename string) *Search {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(file)
	var s Search
	err = dec.Decode(&s)
	if err != nil {
		panic(err)
	}
	return &s
}

// IndexSize returns the term -> Document index size
// for document with ID < maxID
func (s *Search) IndexSize(maxID int) int {
	var indexSize int
	for _, documents := range s.Index {
		if documents[0] <= maxID {
			indexSize++
		}
	}
	return indexSize
}

// TokenSize returns the total number of token in the parsed part of the document
// for document with ID < maxID
func (s *Search) TokenSize(maxID int) int {
	var tokenSize int
	for _, document := range s.Token {
		if document <= maxID {
			tokenSize++
		}
	}
	return tokenSize
}

// CorpusSize returns the total number of document
func (s *Search) CorpusSize() int {
	return s.Size
}

// Serialize a search struct to a file, adding the .gob extension
func (s *Search) Serialize(filename string) {
	file, err := os.Create(filename + ".gob")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	en := gob.NewEncoder(file)
	err = en.Encode(s)
	if err != nil {
		panic(err)
	}
}
