package main

import (
	"encoding/gob"
	"os"
	"strings"
)

// query is a generic interface for query
type query interface {
	execute(*Search) []int
}

type base struct {
	word string
}

// execute returns the query for one word
func (b base) execute(s *Search) []int {
	w := cleanWord(b.word)
	return s.Index[w]
}

type and struct {
	queries []query
}

// execute returns the intersection of two queries
func (a and) execute(s *Search) []int {
	docs := make([]int, 0)
	for _, q := range a.queries {
		newDocs := q.execute(s)
		if len(docs) == 0 {
			docs = newDocs
			continue
		}
		// Perform the merge
		merged := make([]int, 0, len(docs))
		for {
			if len(docs) == 0 || len(newDocs) == 0 {
				break
			}

			if docs[0] == newDocs[0] {
				merged = append(merged, docs[0])
				docs = docs[1:]
				newDocs = newDocs[1:]
			} else if docs[0] < newDocs[0] {
				docs = docs[1:]
			} else {
				newDocs = newDocs[1:]
			}
		}
		docs = merged
	}
	return docs
}

type or struct {
	query1 query
	query2 query
}

func (o or) execute(s *Search) []int {
	docs1 := o.query1.execute(s)
	docs2 := o.query2.execute(s)
	merged := make([]int, 0, len(docs1)+len(docs2))
	for {
		if len(docs1) == 0 {
			merged = append(merged, docs2...)
			break
		}
		if len(docs2) == 0 {
			merged = append(merged, docs1...)
			break
		}

		if docs1[0] == docs2[0] {
			merged = append(merged, docs1[0])
			docs1 = docs1[1:]
			docs2 = docs2[1:]
		} else if docs1[0] < docs2[0] {
			merged = append(merged, docs1[0])
			docs1 = docs1[1:]
		} else {
			merged = append(merged, docs2[0])
			docs2 = docs2[1:]
		}
	}
	return merged
}

// buildQuery build a query from a string, using and operator unless an OR is present
// then only the two words around are considered
func buildQuery(input string) query {
	words := strings.Split(input, " ")
	// we use a and query as a base
	q := and{make([]query, 0, len(words))}
	// we iterate the words slice, building OR couple as needed
	// we remplace words by "" to symbolise that they are used
	for i := range words {
		if strings.ToUpper(words[i]) == "OR" {
			if i == 0 || i == len(words)-1 {
				words[i] = ""
				continue
			}
			// weed out malformaed queries
			if len(words[i-1]) == 0 || len(words[i+1]) == 0 {
				words[i] = ""
				continue
			}
			if words[i-1] == "OR" {
				words[i] = ""
				words[i-1] = ""
				continue
			}
			if words[i+1] == "OR" {
				words[i] = ""
				words[i+1] = ""
				continue
			}
			// add the OR query
			q.queries = append(q.queries, or{base{words[i-1]}, base{words[i+1]}})
			words[i-1] = ""
			words[i+1] = ""
		} else if words[i] != "" {
			q.queries = append(q.queries, base{words[i]})
		}
	}
	return q
}

// Search stores information relevant to parsed documents
type Search struct {
	// Token stores the id of the first document containing a token for heap law
	Token map[string]int
	// Index is a map of token to document ID
	Index map[string][]int
	// Size is the total number of documents
	Size int
	// Titles maps docID to title
	Titles map[int]string
}

func emptySearch() *Search {
	token := make(map[string]int)
	index := make(map[string][]int)
	title := make(map[int]string)
	return &Search{Token: token, Index: index, Titles: title}
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

// Search returns  the list of document title that mention a word
func (s *Search) Search(input string) []string {
	q := buildQuery(input)
	docs := q.execute(s)
	result := make([]string, 0, len(docs))
	for i, doc := range docs {
		// Because result are ordered this prevent printing twice the same doc
		if i == 0 || doc != docs[i-1] {
			result = append(result, s.Titles[doc])
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

// contains check if an int is in a _sorted_ list of int
func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if i == needle {
			return true
		} else if i > needle {
			return false
		}
	}
	return true
}
