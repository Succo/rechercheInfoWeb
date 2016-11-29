package main

import (
	"io"
	"strconv"
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
	s          *Scanner
	field      field
	id         int
	token      map[string]bool
	commonWord map[string]bool
	index      map[string][]int
}

// NewParser creates a parser struct from an io reader and a common word list
func NewParser(r io.Reader, commonWord []string) *Parser {
	index := make(map[string][]int)

	// construct and initialise the common word set
	cw := make(map[string]bool)
	for _, word := range commonWord {
		cw[word] = true
	}

	// construct the token set
	token := make(map[string]bool)

	return &Parser{s: NewScanner(r), commonWord: cw, index: index, token: token}
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
	p.index[lit] = append(p.index[lit], p.id)
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
			p.id, _ = strconv.Atoi(lit)
			return true
		}
		// we add to token to a token set to count it's size
		p.token[lit] = true
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
func (p *Parser) Parse() {
	for p.parse() {
	}
}

// IndexSize returns the terme -> Document index size
func (p *Parser) IndexSize() int {
	return len(p.index)
}

// TokenSize returns the total nulber of token in the parsed part of the document
func (p *Parser) TokenSize() int {
	return len(p.token)
}
