package main

// Document implement a parsed document
type Document struct {
	title string
	// Used to point to the "real document"
	url string
	// store the frequence of keywoard in the document
	freqs map[string]float64
	// stores the total size
	size int
}

func newDocument(title, url string) *Document {
	freqs := make(map[string]float64)
	return &Document{title: title, url: url, freqs: freqs}
}
