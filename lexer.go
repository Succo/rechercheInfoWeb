package main

import (
	"io"
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
	// search only stores result from the indexing
	search *Search
}

// NewCACMParser creates a parser struct from an io reader and a common word list
func NewCACMParser(r io.Reader, commonWord []string) *Parser {
	index := make(map[string][]int)

	// construct and initialise the common word set
	cw := make(map[string]bool)
	for _, word := range commonWord {
		cw[word] = true
	}

	// construct the token set
	token := make(map[string]int)

	search := &Search{token: token, index: index}
	return &Parser{s: NewCACMScanner(r), commonWord: cw, search: search}
}

// NewCS276Parser creates a parser struct from an io reader and a common word list
func NewCS276Parser(root string) *Parser {
	index := make(map[string][]int)

	// construct the common word set
	cw := make(map[string]bool)

	// construct the token set
	token := make(map[string]int)

	search := &Search{token: token, index: index}
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
	p.search.index[lit] = append(p.search.index[lit], p.id)
}

// Parses one "word"
func (p *Parser) parse() bool {
	ch, lit := p.s.Scan()
	if ch == EOF {
		return false
	}
	if ch == Identifiant {
		p.field = identToField(lit)
		return true
	}
	if p.field == other {
		return true
	}
	if ch == WS {
		return true
	}
	if ch == Token {
		// then the only token is the id
		if p.field == id {
			p.id += 1
			return true
		}
		// we store the lowest ID where the word was seen
		// it's easy since id are seen in order
		_, found := p.search.token[lit]
		if !found {
			p.search.token[lit] = p.id
		}
		lit = cleanWord(lit)
		if p.isCommonWord(lit) {
			return true
		}
		// We add non common word to the token list
		p.addWord(lit)
		return true
	}
	return true
}

// Parse will parse the whole buffer
func (p *Parser) Parse() *Search {
	for p.parse() {
	}
	p.search.size = p.id
	return p.search
}
