package main

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// Used to point to the "real document"
	Url string
	// store the frequence of keywoard in the document
	Scores map[string]float64
	// stores the total size
	Size int
	// Tokens is a set of all token
	Tokens map[string]bool
	// pos is the index of the doc in the file
	// only relevant for cacm
	pos int64
}

func newDocument() *Document {
	scores := make(map[string]float64)
	tokens := make(map[string]bool)
	return &Document{Title: "", Url: "", Scores: scores, Tokens: tokens}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	w = cleanWord(w)
	d.Scores[w] += 1
	d.Size += 1
}

// addToken add the token to the set
func (d *Document) addToken(w string) {
	d.Tokens[w] = true
}

// calculTf actually update score to store tf for the term
func (d *Document) calculTf() {
	factor := 1 / float64(d.Size)
	for w, s := range d.Scores {
		d.Scores[w] = s * factor
	}
}
