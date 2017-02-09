package main

import (
	"sort"
	"strings"

	"github.com/surgebase/porter2"
)

// weight serves to identify the different weight that can be used
type weight int

const (
	raw weight = iota
	norm
)

// mergeWithTfIdf calculate the intersection of two list of refs
// It updates the tfidf score in each item
func mergeWithTfIdf(refs1, refs2 []Ref) []Ref {
	intersection := make([]Ref, 0, len(refs1))
	for {
		if len(refs1) == 0 {
			intersection = append(intersection, refs2...)
			break
		}
		if len(refs2) == 0 {
			intersection = append(intersection, refs1...)
			break
		}

		if refs1[0].Id == refs2[0].Id {
			ref := refs1[0]
			ref.RawTfIdf += refs2[0].RawTfIdf
			ref.NormTfIdf += refs2[0].NormTfIdf
			intersection = append(intersection, ref)
			refs1 = refs1[1:]
			refs2 = refs2[1:]
		} else if refs1[0].Id < refs2[0].Id {
			intersection = append(intersection, refs1[0])
			refs1 = refs1[1:]
		} else {
			intersection = append(intersection, refs2[0])
			refs2 = refs2[1:]
		}
	}
	return intersection
}

// VectorQuery effects a vector query on a search object
func VectorQuery(s *Search, input string, w weight) []Ref {
	words := strings.Split(input, " ")
	var results []Ref
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		w = porter2.Stem(w)
		refs := s.Index.get([]byte(w))
		results = mergeWithTfIdf(results, refs)
	}
	if w == raw {
		sort.Sort(rawList(results))
	} else if w == norm {
		sort.Sort(normList(results))
	}
	return results
}

// Define a custom type to add custom method
type rawList []Ref

// Those method satisfy the sort interface
func (r rawList) Len() int      { return len(r) }
func (r rawList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r rawList) Less(i, j int) bool {
	return r[i].RawTfIdf < r[j].RawTfIdf
}

// Define a custom type to add custom method
type normList []Ref

// Those method satisfy the sort interface
func (r normList) Len() int      { return len(r) }
func (r normList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r normList) Less(i, j int) bool {
	return r[i].NormTfIdf < r[j].RawTfIdf
}
