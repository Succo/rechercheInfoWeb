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
	s     Scanner
	field field
	id    int
	// For each token we store the id of the first document where it was seen for heap law
	token      map[string]int
	commonWord map[string]bool
	index      map[string][]int
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

	return &Parser{s: NewCACMScanner(r), commonWord: cw, index: index, token: token}
}

// NewCS276Parser creates a parser struct from an io reader and a common word list
func NewCS276Parser(root string) *Parser {
	index := make(map[string][]int)

	// construct the common word set
	cw := make(map[string]bool)

	// construct the token set
	token := make(map[string]int)

	return &Parser{s: NewCS276Scanner(root), commonWord: cw, index: index, token: token, id: 0}
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
			p.id += 1
			return true
		}
		// we store the lowest ID where the word was seen
		// it's easy since id are seen in order
		_, found := p.token[lit]
		if !found {
			p.token[lit] = p.id
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
func (p *Parser) Parse() {
	for p.parse() {
	}
}

// IndexSize returns the term -> Document index size
// for document with ID < maxID
func (p *Parser) IndexSize(maxID int) int {
	var indexSize int
	for _, documents := range p.index {
		if documents[0] <= maxID {
			indexSize++
		}
	}
	return indexSize
}

// TokenSize returns the total number of token in the parsed part of the document
// for document with ID < maxID
func (p *Parser) TokenSize(maxID int) int {
	var tokenSize int
	for _, document := range p.token {
		if document <= maxID {
			tokenSize++
		}
	}
	return tokenSize
}

// CorpusSize returns the total number of document
func (p *Parser) CorpusSize() int {
	return p.id
}
