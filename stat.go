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

func getStat(s *Search, name string) Stat {
	corpusSize := s.CorpusSize()
	tokenSize := s.TokenSize(corpusSize)
	halfTokenSize := s.TokenSize(corpusSize / 2)
	vocabSize := s.IndexSize(corpusSize)
	halfVocabSize := s.IndexSize(corpusSize / 2)

	// Heaps law calculation
	b := (math.Log(float64(tokenSize)) - math.Log(float64(halfTokenSize))) / (math.Log(float64(vocabSize)) - math.Log(float64(halfVocabSize)))
	k := float64(tokenSize) / (math.Pow(float64(vocabSize), b))

	return Stat{
		Name:       name,
		Documents:  corpusSize,
		Tokens:     tokenSize,
		Vocabulary: vocabSize,
		B:          b,
		K:          k,
	}
}
