package main

import (
	"sort"
	"strings"
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
			ref := Ref{Id: refs1[0].Id, TfIdf: refs1[0].TfIdf + refs2[0].TfIdf}
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
func VectorQuery(s *Search, input string) []Ref {
	words := strings.Split(input, " ")
	var results resultSet
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		w = cleanWord(w)
		refs := s.Index.get([]byte(w))
		results = mergeWithTfIdf(results, refs)
	}
	sort.Sort(results)
	return results
}

// Define a custom type to add custom method
type resultSet []Ref

// Those method satisfy the sort interface
func (r resultSet) Len() int      { return len(r) }
func (r resultSet) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r resultSet) Less(i, j int) bool {
	return r[i].TfIdf < r[j].TfIdf
}
