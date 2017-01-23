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
	return &Parser{s: NewCACMScanner(r), commonWord: cw, search: search}
}

// NewCS276Parser creates a parser struct from an io reader and a common word list
func NewCS276Parser(root string) *Parser {
	// construct the common word set
	// It's empty since CS276 doesn't provide a common word list
	cw := make(map[string]bool)

	search := emptySearch()
	return &Parser{s: NewCS276Scanner(root), commonWord: cw, search: search}
}

// isCommonWord returns wether the word is part of the common word list
// It expects a sorted list, as provide by NewParser
func (p *Parser) isCommonWord(lit string) bool {
	if len(lit) < 3 {
		return true
	}
	_, found := p.commonWord[lit]
	return found
}

// addWord adds the token to the index
func (p *Parser) addWord(lit string) {
	p.doc.addWord(lit)
}

// Parses one "word"
func (p *Parser) parse() bool {
	ch, lit := p.s.Scan()
	switch ch {
	case EOF:
		return false
	case Identifiant:
		p.field = identToField(lit)
		if p.field == id {
			if p.id != 0 {
				// Add the previous document
				p.doc.Title = p.title.String()
				// Add the document
				p.search.AddDocument(p.doc)
			}
			// Reset the document
			p.doc = newDocument(p.id)
			p.title.Reset()
			p.id++
			return true
		}
	case WS:
		if p.field == title {
			p.title.WriteRune(' ')
		}
	case Token:
		// then the only token is the id
		if p.field == title {
			p.title.WriteString(lit)
		}
		// we store the lowest ID where the word was seen
		// it's easy since id are seen in order
		_, found := p.search.Token[lit]
		if !found {
			p.search.Token[lit] = p.id
		}
		lit = cleanWord(lit)
		if p.isCommonWord(lit) {
			return true
		}
		// We add non common word to the token list
		p.addWord(lit)
	}
	return true
}

// Parse will parse the whole buffer
func (p *Parser) Parse() *Search {
	for p.parse() {
	}
	return p.search
}
