package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// field serves to identify the different field
type field int

const (
	id field = iota
	title
	summary
	keyWords
	other
)

func identToField(ident string) field {
	switch ident {
	case ".I":
		return id
	case ".T":
		return title
	case ".W":
		return summary
	case ".K":
		return keyWords
	}
	// This correspond to all untreated field
	return other
}

func cleanWord(word string) string {
	word = strings.ToLower(word)
	return word
}

// ParseCACM creates a cacm scanner, a search struct and connects them
func ParseCACM(r io.Reader, commonWordFile string) *Search {
	commonWord, err := os.Open(commonWordFile)
	if err != nil {
		panic(err)
	}
	defer commonWord.Close()

	cw := make(map[string]bool)
	scanner := bufio.NewScanner(commonWord)
	for scanner.Scan() {
		cw[scanner.Text()] = true
	}

	cacm := NewCACMScanner(r, cw)
	return Parse(cacm)
}

// ParseCS276 creates a parser struct from an io reader and a common word list
func ParseCS276(root string) *Search {
	cs276 := NewCS276Scanner(root)
	return Parse(cs276)
}

// Parse creates a search and populate it with result from a scanner
func Parse(scan Scanner) *Search {
	c := make(chan *Document)
	go scan.Scan(c)
	search := emptySearch()
	for doc := range c {
		search.AddDocument(doc)
	}
	return search
}
