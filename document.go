package main

import (
	"math"

	"github.com/surgebase/porter2"
)

// weight serves to identify the different weight that can be used
type weight int

const (
	raw weight = iota
	norm
	half
	total int = iota // serves as a counter
)

// weightName can be used to iterate over the differentes weightFun
// makes adding new weight function easy
var weightName = [...]string{
	"raw frequency",
	"log normalization",
	"double normalization 0.5",
}

// weight is a fixed size array where the different tfidif for one term values can be stored
type weights [total]float64

func zip(w1, w2 weights) (w weights) {
	w[raw] = w1[raw] + w2[raw]
	w[norm] = w1[norm] + w2[norm]
	w[half] = w1[half] + w2[half]
	return w
}

func scale(w weights, c float64) (newW weights) {
	newW[raw] = c * w[raw]
	newW[norm] = c * w[norm]
	newW[half] = c * w[half]
	return newW
}

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// store the count of each term in the document
	Count map[string]int
	// store the different frequence of a term in the document
	Scores map[string]weights
	// stores the total size
	Size int
	// Tokens counts the number of token
	Tokens int
	// Id is the id of the document (unique in the search)
	Id uint
}

func newDocument() *Document {
	scores := make(map[string]weights)
	count := make(map[string]int)
	return &Document{Scores: scores, Count: count}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	if len(w) > 3 {
		w = porter2.Stem(w)
	}
	d.Count[w]++
	d.Size += 1
}

// addToken add the token to the set
func (d *Document) addToken(w string) {
	d.Tokens++
}

// calculScore actually update score to store tf for the term
func (d *Document) calculScore() {
	// Get the maximun tf
	var max int
	for _, s := range d.Count {
		if s > max {
			max = s
		}
	}
	maxF := 1 / float64(max)
	var score weights
	for w, s := range d.Count {
		tf := float64(s)
		score[raw] = tf
		score[norm] = 1 + math.Log(tf)
		score[half] = 0.5 + 0.5*tf*maxF
		d.Scores[w] = score
	}
}
