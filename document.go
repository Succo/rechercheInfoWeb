package main

import (
	"math"

	"github.com/surgebase/porter2"
)

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// store the frequence of keyword in the document
	RawScores map[string]float64
	// store the log normalized frequence of keyword in the document
	NormScores map[string]float64
	// stores the total size
	Size int
	// Tokens counts the number of token
	Tokens int
	// pos is the index of the doc in the file
	// only relevant for cacm
	pos int64
}

func newDocument() *Document {
	raw := make(map[string]float64)
	norm := make(map[string]float64)
	return &Document{RawScores: raw, NormScores: norm}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	w = porter2.Stem(w)
	d.RawScores[w] += 1
	d.Size += 1
}

// addToken add the token to the set
func (d *Document) addToken(w string) {
	d.Tokens++
}

// calculScore actually update score to store tf for the term
func (d *Document) calculScore() {
	factor := 1 / float64(d.Size)
	for w, s := range d.RawScores {
		tf := s * factor
		d.RawScores[w] = tf
		d.NormScores[w] = 1 + math.Log(tf)
	}
}
