package main

import "strings"

// query is a generic interface for query
type query interface {
	execute(*Search) []*Document
}

type base struct {
	word string
}

// execute returns the query for one word
func (b base) execute(s *Search) []*Document {
	w := cleanWord(b.word)
	return s.Index[w]
}

type and struct {
	queries []query
}

// execute returns the intersection of two queries
func (a and) execute(s *Search) []*Document {
	docs := make([]*Document, 0)
	for _, q := range a.queries {
		newDocs := q.execute(s)
		if len(docs) == 0 {
			docs = newDocs
			continue
		}
		// Perform the merge
		merged := make([]*Document, 0, len(docs))
		for {
			if len(docs) == 0 || len(newDocs) == 0 {
				break
			}

			if docs[0] == newDocs[0] {
				merged = append(merged, docs[0])
				docs = docs[1:]
				newDocs = newDocs[1:]
			} else if docs[0].Id < newDocs[0].Id {
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

func (o or) execute(s *Search) []*Document {
	docs1 := o.query1.execute(s)
	docs2 := o.query2.execute(s)
	merged := make([]*Document, 0, len(docs1)+len(docs2))
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
		} else if docs1[0].Id < docs2[0].Id {
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
