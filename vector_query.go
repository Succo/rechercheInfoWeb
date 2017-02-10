package main

import (
	"sort"
	"strings"

	"github.com/surgebase/porter2"
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
			ref.Weights = zip(ref.Weights, refs2[0].Weights)
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
func VectorQuery(s *Search, input string, wf weight) []Ref {
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
