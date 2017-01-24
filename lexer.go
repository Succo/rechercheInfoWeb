package main

import (
	"bufio"
	"bytes"
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

// Parser is the struct that will parse using a Scanner
// and hold the parsed data
type Parser struct {
	// Those are the field used while parsing
	s          Scanner
	field      field
	id         int
	commonWord map[string]bool
	title      bytes.Buffer
	doc        *Document
	// search only stores result from the indexing
	search *Search
}

// NewCACMParser creates a parser struct from an io reader and a common word list
func NewCACMParser(r io.Reader, commonWordFile string) *Parser {
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

	search := emptySearch()
	return &Parser{s: NewCACMScanner(r, cw), commonWord: cw, search: search}
}

// NewCS276Parser creates a parser struct from an io reader and a common word list
func NewCS276Parser(root string) *Parser {
	// construct the common word set
	// It's empty since CS276 doesn't provide a common word list
	cw := make(map[string]bool)

	search := emptySearch()
	return &Parser{s: NewCS276Scanner(root), commonWord: cw, search: search}
}

// addWord adds the token to the index
func (p *Parser) addWord(lit string) {
	p.doc.addWord(lit)
}

// Parse will parse the whole buffer
func (p *Parser) Parse() *Search {
	c := make(chan *Document)
	go p.s.Scan(c)
	for doc := range c {
		p.search.AddDocument(doc)
	}
	return p.search
}
