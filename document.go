package main

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// store the frequence of keywoard in the document
	Scores map[string]float64
	// stores the total size
	Size int
	// Tokens counts the number of token
	Tokens int
	// pos is the index of the doc in the file
	// only relevant for cacm
	pos int64
}

func newDocument() *Document {
	scores := make(map[string]float64)
	return &Document{Scores: scores}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	w = cleanWord(w)
	d.Scores[w] += 1
	d.Size += 1
}

// addToken add the token to the set
func (d *Document) addToken(w string) {
	d.Tokens++
}

// calculTf actually update score to store tf for the term
func (d *Document) calculTf() {
	factor := 1 / float64(d.Size)
	for w, s := range d.Scores {
		d.Scores[w] = s * factor
	}
}
