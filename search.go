package main

import (
	"encoding/gob"
	"os"
)

// Search stores information relevant to parsed documents
type Search struct {
	// Token stores the id of the first document containing a token for heap law
	Token map[string]int
	// Index is a map of token document pointers
	Index map[string][]*Document
	// Size is the total number of documents
	Size int
}

func emptySearch() *Search {
	token := make(map[string]int)
	index := make(map[string][]*Document)
	return &Search{Token: token, Index: index}
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

func (s *Search) AddDocument(d *Document) {
	d.calculFreqs()
	for w, _ := range d.Freqs {
		s.Index[w] = append(s.Index[w], d)
	}
}

// IndexSize returns the term -> Document index size
// for document with ID < maxID
func (s *Search) IndexSize(maxID int) int {
	var indexSize int
	for _, documents := range s.Index {
		if documents[0].Id <= maxID {
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

// Search returns  the list of document title that mention a word
func (s *Search) Search(input string) []string {
	q := buildQuery(input)
	docs := q.execute(s)
	result := make([]string, 0, len(docs))
	for i, doc := range docs {
		// Because result are ordered this prevent printing twice the same doc
		if i == 0 || doc != docs[i-1] {
			result = append(result, doc.Title)
		}
	}
	return result
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
