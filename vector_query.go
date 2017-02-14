package main

import (
	"sort"
	"strings"
	"unicode"

	"github.com/gonum/floats"
	"github.com/surgebase/porter2"
)

// splitter is a rules by which words are splitted
// Basically only keep letters and number
func splitter(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

// mergeWithTfIdf calculate the merge of a sorted list of documents
// calculating the norm in the sametime
func mergeWithTfIdf(documents [][]Ref, wf weight) []Ref {
	merge := make([]Ref, 0, len(documents[0]))
	// Temporaty slice to store result
	temps := make([]float64, 0, len(documents))
	var min int
	for {
		temps = temps[:0]
		// Find the lowest Id
		min = -1
		for _, refs := range documents {
			if len(refs) > 0 {
				if min == -1 || refs[0].Id < min {
					min = refs[0].Id
				}
			}
		}
		// No document anymore
		if min == -1 {
			break
		}
		for i, refs := range documents {
			if len(refs) != 0 && refs[0].Id == min {
				temps = append(temps, refs[0].Weights[wf])
				documents[i] = refs[1:]
			}
		}
		ref := Ref{
			Id: min,
		}
		ref.Weights[wf] = floats.Sum(temps) / floats.Norm(temps, 2)
		merge = append(merge, ref)
	}
	return merge
}

// VectorQuery effects a vector query on a search object
func VectorQuery(s *Search, input string, wf weight) []Ref {
	words := strings.FieldsFunc(input, splitter)
	documents := make([][]Ref, len(words))
	for i, w := range words {
		if s.CW[w] {
			continue
		}
		if len(w) > 3 {
			w = porter2.Stem(w)
		}
		documents[i] = s.Index.get([]byte(w))
	}
	results := mergeWithTfIdf(documents, wf)
	if wf == raw {
		sort.Sort(rawList(results))
	} else if wf == norm {
		sort.Sort(normList(results))
	} else if wf == half {
		sort.Sort(halfList(results))
	}
	return results
}

// Define a custom type to add custom method
type rawList []Ref

// Those method satisfy the sort interface
func (r rawList) Len() int      { return len(r) }
func (r rawList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r rawList) Less(i, j int) bool {
	return r[i].Weights[raw] < r[j].Weights[raw]
}

// Define a custom type to add custom method
type normList []Ref

// Those method satisfy the sort interface
func (r normList) Len() int      { return len(r) }
func (r normList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r normList) Less(i, j int) bool {
	return r[i].Weights[norm] < r[j].Weights[norm]
}

// Define a custom type to add custom method
type halfList []Ref

// Those method satisfy the sort interface
func (r halfList) Len() int      { return len(r) }
func (r halfList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r halfList) Less(i, j int) bool {
	return r[i].Weights[half] < r[j].Weights[half]
}
