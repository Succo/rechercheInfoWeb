package main

import "strings"

// query is a generic interface for query
type query interface {
	execute(*Search) []Ref
}

type base struct {
	word string
}

// execute returns the query for one word
func (b base) execute(s *Search) []Ref {
	w := cleanWord(b.word)
	return s.Index[w]
}

type and struct {
	queries []query
}

// execute returns the intersection of two queries
func (a and) execute(s *Search) []Ref {
	var refs []Ref
	for _, q := range a.queries {
		newRefs := q.execute(s)
		if len(refs) == 0 {
			refs = newRefs
			continue
		}
		// Perform the merge
		merged := make([]Ref, 0, len(refs))
		for {
			if len(refs) == 0 || len(newRefs) == 0 {
				break
			}

			if refs[0] == newRefs[0] {
				merged = append(merged, refs[0])
				refs = refs[1:]
				newRefs = newRefs[1:]
			} else if refs[0].Id < newRefs[0].Id {
				refs = refs[1:]
			} else {
				newRefs = newRefs[1:]
			}
		}
		refs = merged
	}
	return refs
}

type or struct {
	query1 query
	query2 query
}

func (o or) execute(s *Search) []Ref {
	refs1 := o.query1.execute(s)
	refs2 := o.query2.execute(s)
	merged := make([]Ref, 0, len(refs1)+len(refs2))
	for {
		if len(refs1) == 0 {
			merged = append(merged, refs2...)
			break
		}
		if len(refs2) == 0 {
			merged = append(merged, refs1...)
			break
		}

		if refs1[0] == refs2[0] {
			merged = append(merged, refs1[0])
			refs1 = refs1[1:]
			refs2 = refs2[1:]
		} else if refs1[0].Id < refs2[0].Id {
			merged = append(merged, refs1[0])
			refs1 = refs1[1:]
		} else {
			merged = append(merged, refs2[0])
			refs2 = refs2[1:]
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
