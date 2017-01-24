package main

import (
	"encoding/gob"
	"os"
)

// Result is a document as returned by a Search
type Result struct {
	Name string
	Url  string
}

// Ref is a reference to a document
type Ref struct {
	Id   int
	Freq float64
}

// Search stores information relevant to parsed documents
type Search struct {
	// Token stores the id of the first document containing a token for heap law
	Token map[string]int
	// Index is a map of token document pointers
	Index map[string][]Ref
	// Size is the total number of documents
	Size int
	// Titles stores document title
	Titles []string
	// Url stores url to document
	Urls []string
}

func emptySearch() *Search {
	token := make(map[string]int)
	index := make(map[string][]Ref)
	var titles []string
	return &Search{Token: token, Index: index, Titles: titles}
}

// NewSearch generates a Search loading a serialized file
func NewSearch(filename string) *Search {
	file, err := os.Open(filename + ".gob")
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

// AddDocument adds a parsed document to it's indexes
func (s *Search) AddDocument(d *Document) {
	for w, f := range d.Freqs {
		s.Index[w] = append(s.Index[w], Ref{s.Size, f})
	}
	for t := range d.Tokens {
		_, found := s.Token[t]
		if !found {
			s.Token[t] = s.Size
		}
	}
	s.Size++
	s.Titles = append(s.Titles, d.Title)
	s.Urls = append(s.Urls, d.Url)
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
func (s *Search) Search(input string) []Result {
	q := buildQuery(input)
	refs := q.execute(s)
	results := make([]Result, 0, len(refs))
	for i, ref := range refs {
		// Because result are ordered this prevent printing twice the same doc
		if i == 0 || ref.Id != refs[i-1].Id {
			results = append(results, Result{s.Titles[ref.Id], s.Urls[ref.Id]})
		}
	}
	return results
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
