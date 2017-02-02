package main

import (
	"sync"

	porterstemmer "github.com/reiver/go-porterstemmer"
)

type Stemmer struct {
	lock  sync.Mutex
	cache map[string]string
}

func NewStemmer() *Stemmer {
	cache := make(map[string]string)
	return &Stemmer{cache: cache}
}

func (s *Stemmer) stem(w string) string {
	s.lock.Lock()
	stem, found := s.cache[w]
	if !found {
		stem := cleanWord(w)
		s.cache[w] = stem
		s.lock.Unlock()
		return stem
	} else {
		s.lock.Unlock()
		return stem
	}
}

func cleanWord(word string) string {
	stem := porterstemmer.StemString(word)
	return stem
}
