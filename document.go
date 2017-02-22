package main

import "github.com/surgebase/porter2"

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

func scale(w *weights, c float64) {
	w[raw] = c * w[raw]
	w[norm] = c * w[norm]
	w[half] = c * w[half]
}

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// store the count of each term in the document
	// Words is an ordered list of lexemes
	// Count is the number of occurence for the word at the same index
	Count []int
	Words []string
	// stores the total size
	Size int
	// Tokens counts the number of token
	Tokens int
	// Id is the id of the document (unique in the search)
	Id int
}

func newDocument() *Document {
	return &Document{}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	if len(w) > 3 {
		w = porter2.Stem(w)
	}
	i := getWordIndex(d.Words, w)
	if i < len(d.Words) && d.Words[i] == w {
		d.Count[i]++
	} else if i == len(d.Words) {
		d.Count = append(d.Count, 1)
		d.Words = append(d.Words, w)
	} else {
		d.Count = append(d.Count, 0)
		d.Words = append(d.Words, "")
		copy(d.Count[i+1:], d.Count[i:])
		copy(d.Words[i+1:], d.Words[i:])
		d.Count[i]++
		d.Words[i] = w
	}
	d.Size += 1
}

// addToken add the token to the set
func (d *Document) addToken(w string) {
	d.Tokens++
}

func (d *Document) reset() {
	d.Count = d.Count[:0]
	d.Words = d.Words[:0]
	d.Size = 0
	d.Tokens = 0
}

func getWordIndex(words []string, w string) int {
	switch {
	case len(words) == 0:
		return 0

	case len(words) < 10:
		for i, word := range words {
			if word >= w {
				return i
			}
		}
		return len(words)

	default:
		// If the slice is long use binary search
		// better tuning needed here
		min := 0
		max := len(words)
		var match int
		for min < max {
			match = min + (max-min)/2
			if words[match] < w {
				min = match + 1
			} else {
				max = match
			}
		}
		return min
	}

}
