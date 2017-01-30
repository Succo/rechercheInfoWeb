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
	Id    int
	TfIdf float64
}

// Search stores information relevant to parsed documents
type Search struct {
	Retriever
	Stat   Stat
	Corpus string
	// Token stores the id of the first document containing a token for heap law
	Token map[string]int
	// Index is a trie of token document pointers
	Index *Node
	// Size is the total number of documents
	Size int
	// Titles stores document title
	Titles []string
	// Url stores url to document
	Urls []string
}

func emptySearch(corpus string) *Search {
	token := make(map[string]int)
	index := NewTrie()
	var titles []string
	return &Search{Token: token, Index: index, Titles: titles, Corpus: corpus}
}

// AddDocMetaData adds a parsed document metadata
func (s *Search) AddDocMetaData(d *Document) {
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
	return s.Index.getInfIndex(maxID)
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
	refs := query(s, input)
	results := make([]Result, 0, len(refs))
	for i, ref := range refs {
		// Because result are ordered this prevent printing twice the same doc
		if i == 0 || ref.Id != refs[i-1].Id {
			results = append(results, Result{s.Titles[ref.Id], s.Urls[ref.Id]})
		}
	}
	return results
}

// Serialize a search struct to a file
// we only serialize the index, the titles and the urls list
// no need to consider the tokens since they only serve to calculate HEAP law
func (s *Search) Serialize() {
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

	urls, err := os.Create("indexes/" + s.Corpus + ".urls")
	if err != nil {
		panic(err)
	}
	defer urls.Close()
	en = gob.NewEncoder(urls)
	err = en.Encode(s.Urls)
	if err != nil {
		panic(err)
	}
	urls.Sync()
	urls.Close()

	index, err := os.Create("indexes/" + s.Corpus + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	en = gob.NewEncoder(index)
	err = en.Encode(s.Index)
	if err != nil {
		panic(err)
	}
	index.Sync()
	index.Close()

	stat, err := os.Create("indexes/" + s.Corpus + ".stat")
	if err != nil {
		panic(err)
	}
	defer stat.Close()
	en = gob.NewEncoder(stat)
	err = en.Encode(s.Stat)
	if err != nil {
		panic(err)
	}
	stat.Sync()
	stat.Close()

	s.Retriever.Serialize(s.Corpus)
}

// Unserialize reloads what's needed from disk
func Unserialize(name string) *Search {
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

	urls, err := os.Open("indexes/" + name + ".urls")
	if err != nil {
		panic(err)
	}
	defer urls.Close()
	en = gob.NewDecoder(urls)
	err = en.Decode(&s.Urls)
	if err != nil {
		panic(err)
	}
	urls.Close()

	index, err := os.Open("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	en = gob.NewDecoder(index)
	err = en.Decode(&s.Index)
	if err != nil {
		panic(err)
	}
	index.Close()

	stat, err := os.Open("indexes/" + name + ".stat")
	if err != nil {
		panic(err)
	}
	defer stat.Close()
	en = gob.NewDecoder(stat)
	err = en.Decode(&s.Stat)
	if err != nil {
		panic(err)
	}
	stat.Close()

	return s
}
