// Stat are all structure returning information and statistics about search type
package main

import "math"

type Stat struct {
	Name       string
	Documents  int
	Vocabulary int
	Tokens     int
	B          float64
	K          float64
}

func getStat(s *Search) Stat {
	length := s.TokenSize(s.Size)
	halfLength := s.TokenSize(s.Size / 2)
	distinct := s.IndexSize(s.Size)
	halfDistinct := s.IndexSize(s.Size / 2)

	// Heaps law calculation
	b := (math.Log(float64(distinct)) - math.Log(float64(halfDistinct))) / (math.Log(float64(length)) - math.Log(float64(halfLength)))
	k := float64(distinct) / (math.Pow(float64(length), b))

	return Stat{
		Name:       s.Corpus,
		Documents:  s.Size,
		Tokens:     length,
		Vocabulary: distinct,
		B:          b,
		K:          k,
	}
}
