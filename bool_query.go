// Bool_query implements the boolean queery function
//
// it uses Shunting-yard algorithm to parse the query
// it tries to be smart when dealing with NOT operator but will faill silently if it can't
// and default to AND operator if not specified (for example "test student" == "test AND student")
package main

import (
	"log"
	"strings"

	"github.com/surgebase/porter2"
)

type operator int

const (
	// The operators are defined in their precedence order
	and operator = iota
	or
	not
	leftParen
)

// stack is a basic stack implementation
type stack []operator

func (s stack) push(t operator) stack {
	return append(s, t)
}

func (s stack) pop() (stack, operator) {
	l := len(s)
	return s[:l-1], s[l-1]
}

// BQuery is the interface for all boolean query
// prec is the previous result, only used by the not operator
// because it doesn't wan't to be executed on empty question
// it would generate too big of a result set
type BQuery interface {
	evaluate(s *Search, prec []Ref) []Ref
	isNot() bool
}

// WordQuery implements the boolean query interface
// and correspond to a single word query
type WordQuery struct {
	w string
}

func (w WordQuery) evaluate(s *Search, prec []Ref) []Ref {
	if len(w.w) > 3 {
		w.w = porter2.Stem(w.w)
	}
	return s.Index.get(w.w)
}

func (w WordQuery) isNot() bool { return false }

// NotQuery implements the negation of a query
// it will returns empty if not applied on an alread defined set
type NotQuery struct {
	b BQuery
}

func (n NotQuery) evaluate(s *Search, prec []Ref) []Ref {
	refs := n.b.evaluate(s, []Ref{})
	return remove(prec, refs)
}

func (n NotQuery) isNot() bool { return true }

// OrQuery implements the union of two queries
type OrQuery struct {
	b1 BQuery
	b2 BQuery
}

func (o OrQuery) evaluate(s *Search, prec []Ref) []Ref {
	res1 := o.b1.evaluate(s, prec)
	res2 := o.b2.evaluate(s, prec)
	return union(res1, res2)
}

func (o OrQuery) isNot() bool { return false }

// AndQuery is implements the  intersection of two queries
// it tries to keep not element for the end
type AndQuery struct {
	b1 BQuery
	b2 BQuery
}

func (a AndQuery) evaluate(s *Search, prec []Ref) []Ref {
	if a.b1.isNot() && !a.b2.isNot() {
		a.b1, a.b2 = a.b2, a.b1
	}
	res1 := a.b1.evaluate(s, prec)
	res2 := a.b2.evaluate(s, res1)
	return intersect(res1, res2)
}

func (a AndQuery) isNot() bool { return false }

func intersect(refs1, refs2 []Ref) []Ref {
	intersection := make([]Ref, 0, len(refs1))
	for {
		if len(refs1) == 0 || len(refs2) == 0 {
			break
		}

		if refs1[0].Id == refs2[0].Id {
			intersection = append(intersection, refs1[0])
			refs1 = refs1[1:]
			refs2 = refs2[1:]
		} else if refs1[0].Id < refs2[0].Id {
			refs1 = refs1[1:]
		} else {
			refs2 = refs2[1:]
		}
	}
	return intersection
}

func union(refs1, refs2 []Ref) []Ref {
	union := make([]Ref, 0, len(refs1)+len(refs2))
	for {
		if len(refs1) == 0 {
			union = append(union, refs2...)
			break
		}
		if len(refs2) == 0 {
			union = append(union, refs1...)
			break
		}
		if refs1[0].Id == refs2[0].Id {
			union = append(union, refs1[0])
			refs1 = refs1[1:]
			refs2 = refs2[1:]
		} else if refs1[0].Id < refs2[0].Id {
			union = append(union, refs1[0])
			refs1 = refs1[1:]
		} else {
			union = append(union, refs2[0])
			refs2 = refs2[1:]
		}
	}
	return union
}

// remove removes element of refs2 from refs1
func remove(refs1, refs2 []Ref) []Ref {
	removed := make([]Ref, 0, len(refs1))
	for {
		if len(refs1) == 0 {
			return removed
		}
		if len(refs2) == 0 {
			removed = append(removed, refs1...)
			break
		}
		if refs1[0].Id == refs2[0].Id {
			refs1 = refs1[1:]
			refs2 = refs2[1:]
		} else if refs1[0].Id < refs2[0].Id {
			removed = append(removed, refs1[0])
			refs1 = refs1[1:]
		} else {
			refs2 = refs2[1:]
		}
	}
	return removed
}

// BooleanQuery parses a query tring it's best to interpret it as a boolean query
// it will fail silently and returns empty results if it can't
func BooleanQuery(s *Search, input string) (results []Ref) {
	// split the words == really basic parsing of the query
	words := strings.FieldsFunc(input, splitter)

	// query interpretation using Shunting-yard
	// query is the output queue
	var query []BQuery
	// operators is the stack of operator
	operators := stack(make([]operator, 0))
	for i, word := range words {
		switch strings.ToUpper(word) {
		case "OR", "AND", "NOT":
			var op operator
			switch strings.ToUpper(word) {
			case "OR":
				op = or
			case "AND":
				op = and
			case "NOT":
				op = not
			}
			// first treat all operator with a higher precedence
			for len(operators) > 0 {
				var oldOp operator
				operators, oldOp = operators.pop()
				if op < oldOp {
					query = addBOperator(query, oldOp)
				} else {
					operators = operators.push(oldOp)
					break
				}
			}
			// then add the operator to the stack
			operators = operators.push(op)
		case "(":
			// just add it to the stack
			operators = operators.push(leftParen)
		case ")":
			// pop from the stack until the matching parentheses is found
			for len(operators) > 0 {
				var oldOp operator
				operators, oldOp = operators.pop()
				if oldOp != leftParen {
					query = addBOperator(query, oldOp)
				} else {
					break
				}
			}
		default:
			query = append(query, WordQuery{w: word})
			// default "OR" operaor between words
			if i+1 < len(words) {
				switch strings.ToUpper(words[i+1]) {
				case "OR", "AND", "(", ")":
					// An operator is already present
					// Do nothing
					continue
				default:
					// Add an or operator
					// repeat the same insertion procedure from the previous case
					op := and
					for len(operators) > 0 {
						operators, oldOp := operators.pop()
						if oldOp < op {
							query = addBOperator(query, oldOp)
						} else {
							operators = operators.push(oldOp)
							break
						}
					}
					operators = operators.push(op)
				}
			}
		}
	}
	// empty the stack of operator
	for i := len(operators) - 1; i >= 0; i-- {
		query = addBOperator(query, operators[i])
	}
	if len(query) != 1 {
		log.Println("error when processing bool query")
	} else {
		results = query[0].evaluate(s, make([]Ref, 0))
	}
	return results
}

func addBOperator(out []BQuery, op operator) []BQuery {
	// case wher it's an unary operator
	if op == not {
		l := len(out)
		if l < 1 {
			// that would mean two operator in a row
			// Don't throw error just silently "fix" the query
			return out
		}
		b := out[l-1]
		out = out[:l-1]
		out = append(out, NotQuery{b})
		return out
	}

	l := len(out)
	if l < 2 {
		// that would mean two operator in a row
		// Don't throw error just silently "fix" the query
		return out
	}
	// Apply the operator to the two BExpr in out
	b1 := out[l-1]
	b2 := out[l-2]
	out = out[:l-2]
	switch op {
	case or:
		out = append(out, OrQuery{b1, b2})
	case and:
		out = append(out, AndQuery{b1, b2})
	}
	return out
}
