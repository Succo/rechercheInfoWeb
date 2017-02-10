package main

import (
	"encoding/gob"
	"os"
	"time"
)

// Result is a document as returned by a Search
type Result struct {
	Name string
	Url  string
}

// Ref is a reference to a document
type Ref struct {
	Id      int
	Weights weights
}

// Search stores information relevant to parsed documents
type Search struct {
	Stat   Stat
	Perf   Perf
	Corpus string
	// Tokens stores the number of token for each document
	// Only used for heaps law so it's no serialized
	Tokens []int
	// Index is a trie of token document pointers
	Index *Root
	// Size is the total number of documents
	Size int
	// Titles stores document title
	Titles []string
	// toUrl generates URL from id and title, the function depends of the corpus
	toUrl func(int, string) string
}

func emptySearch(corpus string) *Search {
	return &Search{Corpus: corpus}
}

// AddDocMetaData adds a parsed document metadata
func (s *Search) AddDocMetaData(d *Document) {
	s.Tokens = append(s.Tokens, d.Tokens)
	s.Titles = append(s.Titles, d.Title)
}

// IndexSize returns the term -> Document index size
// for document with ID < maxID
func (s *Search) IndexSize(maxID int) int {
	return s.Index.getInfIndex(maxID)
}

// TokenSize returns the total number of token
// for document with ID < maxID
func (s *Search) TokenSize(maxID int) int {
	var size int
	for _, toks := range s.Tokens[:maxID] {
		size += toks
	}
	return size
}

// BooleanSeach performs a Boolean search based on a query
func (s *Search) BooleanSearch(input string) []Result {
	refs := BooleanQuery(s, input)
	return s.refToResult(refs)
}

// VectorSearch performs a Vectorial search using TfIdf scores
func (s *Search) VectorSearch(input string, w weight) []Result {
	refs := VectorQuery(s, input, w)
	return s.refToResult(refs)
}

// refToResult transform a list of ref in a list of printable result
// i.e remplace docID by doc metadata
func (s *Search) refToResult(refs []Ref) []Result {
	results := make([]Result, 0, len(refs))
	for i, ref := range refs {
		// Because result are ordered this prevent printing twice the same doc
		if i == 0 || ref.Id != refs[i-1].Id {
			results = append(results,
				Result{s.Titles[ref.Id], s.toUrl(ref.Id, s.Titles[ref.Id])})
		}
	}
	return results
}

// Serialize a search struct to a file
// we only serialize the index, the titles and the urls list
// no need to consider the tokens since they only serve to calculate HEAP law
func (s *Search) Serialize() {
	now := time.Now()
	titles, err := os.Create("indexes/" + s.Corpus + ".titles")
	if err != nil {
		panic(err)
	}
	defer titles.Close()
	en := gob.NewEncoder(titles)
	err = en.Encode(s.Titles)
	if err != nil {
		panic(err)
	}
	titles.Sync()
	titles.Close()

	s.Index.Serialize(s.Corpus)
	s.Perf.Serialization = time.Since(now)
	s.Perf = s.Perf.getFinalValues()

	meta, err := os.Create("indexes/" + s.Corpus + ".meta")
	if err != nil {
		panic(err)
	}
	defer meta.Close()
	en = gob.NewEncoder(meta)
	err = en.Encode(s.Stat)
	if err != nil {
		panic(err)
	}
	err = en.Encode(s.Perf)
	if err != nil {
		panic(err)
	}
	meta.Sync()
	meta.Close()
}

// UnserializeSearch reloads what's needed from disk
func UnserializeSearch(name string) *Search {
	s := &Search{}
	s.Corpus = name
	titles, err := os.Open("indexes/" + name + ".titles")
	if err != nil {
		panic(err)
	}
	defer titles.Close()
	en := gob.NewDecoder(titles)
	err = en.Decode(&s.Titles)
	if err != nil {
		panic(err)
	}
	titles.Close()

	meta, err := os.Open("indexes/" + name + ".meta")
	if err != nil {
		panic(err)
	}
	defer meta.Close()
	en = gob.NewDecoder(meta)
	err = en.Decode(&s.Stat)
	if err != nil {
		panic(err)
	}
	err = en.Decode(&s.Perf)
	if err != nil {
		panic(err)
	}
	meta.Close()

	s.Index = UnserializeTrie(name)
	return s
}
