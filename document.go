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
	Id   int
}

func newDocument() *Document {
	freqs := make(map[string]float64)
	return &Document{Title: "", Url: "", Freqs: freqs}
}

// addWord add a word to the model, for now freqs are only stored as count actually
func (d *Document) addWord(w string) {
	d.Freqs[w] += 1
	d.Size += 1
}

// calculFreqs really calculate the frequenciez
func (d *Document) calculFreqs() {
	size := float64(d.Size)
	for w, freq := range d.Freqs {
		d.Freqs[w] = freq / size
	}
}
