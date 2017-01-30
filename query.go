package main

import "strings"

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
		if refs1[0] == refs2[0] {
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
		if refs1[0] == refs2[0] {
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

func query(s *Search, input string) []Ref {
	words := strings.Split(input, " ")
	var results []Ref
	for i := 0; i < len(words); i++ {
		switch {
		case len(words[i]) == 0:
		case i < len(words)-1 && strings.ToUpper(words[i+1]) == "OR":
			if i >= len(words)-2 {
				return results
			}
			word1 := cleanWord(words[i])
			word2 := cleanWord(words[i+2])
			refs1 := s.Index.get([]byte(word1))
			refs2 := s.Index.get([]byte(word2))
			if i == 0 {
				results = union(refs1, refs2)
			} else {
				results = intersect(results, union(refs1, refs2))
			}
			i += 2 // Jump two words
		case strings.ToUpper(words[i]) == "NOT":
			if i == len(words)-1 {
				return results
			}
			word := cleanWord(words[i+1])
			not := s.Index.get([]byte(word))
			results = remove(results, not)
			i++
		default:
			word := cleanWord(words[i])
			refs := s.Index.get([]byte(word))
			if i == 0 {
				results = refs
			} else {
				results = intersect(results, refs)
			}
		}
	}
	return results

}
