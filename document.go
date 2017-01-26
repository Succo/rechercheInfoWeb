package main

// Document implement a parsed document
// It's a temporary structure until frequences are calulated
type Document struct {
	Title string
	// Used to point to the "real document"
	Url string
	// store the frequence of keywoard in the document
	Freqs map[string]float64
	// stores the total size
	Size int
	// Tokens is a set of all token
	Tokens map[string]bool
	// pos is the index of the doc in the file
	// only relevant for cacm
	pos int64
}

func newDocument() *Document {
	freqs := make(map[string]float64)
	tokens := make(map[string]bool)
	return &Document{Title: "", Url: "", Freqs: freqs, Tokens: tokens}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	w = cleanWord(w)
	d.Freqs[w] += 1
	d.Size += 1
}

func (d *Document) addToken(w string) {
	d.Tokens[w] = true
}

// calculFreqs really calculate the frequenciez
func (d *Document) calculFreqs() {
	factor := 1 / float64(d.Size)
	for w, freq := range d.Freqs {
		d.Freqs[w] = freq * factor
	}
}
