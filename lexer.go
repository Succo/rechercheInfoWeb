package main

import (
	"io"
	"sort"
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

type Parser struct {
	s          *Scanner
	field      field
	id         int
	commonWord []string
	index      map[string][]int
}

// NewParser creates a parser struct from an io reader and a common word list
func NewParser(r io.Reader, commonWord []string) *Parser {
	sort.Strings(commonWord)
	index := make(map[string][]int)
	return &Parser{s: NewScanner(r), commonWord: commonWord, index: index}
}

// isCommonWord returns wether the word is part of the common word list
// It expeccts a sorted list, as provide by NewParser
func (p *Parser) isCommonWord(lit string) bool {
	for _, word := range p.commonWord {
		if lit == word {
			return true
		}
		if lit > word {
			return false
		}
	}
	return false
}

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
	if ch == WS {
		return true
	}
	if ch == Token {
		// then the only token is the id
		if p.field == id {
			p.id = int(id)
			return true
		}
		if p.isCommonWord(lit) {
			return true
		}
		p.addWord(lit)
		return true
	}
	return true
}

func (p *Parser) Parse() {
	for {
		if !p.parse() {
			break
		}
	}
}
