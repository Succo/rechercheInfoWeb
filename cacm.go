package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
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

// CACMScanner will walk the buffer and return document one by one
type CACMScanner struct {
	r          *bufio.Reader
	field      field
	title      bytes.Buffer
	commonWord map[string]bool
	doc        *Document
	id         int
}

// NewCACMScanner create a CACMScanner from an io reader
func NewCACMScanner(r io.Reader, cw map[string]bool) *CACMScanner {
	return &CACMScanner{r: bufio.NewReader(r), commonWord: cw}
}

func (s *CACMScanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *CACMScanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *CACMScanner) peek() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		ch = eof
	}
	s.r.UnreadRune()
	return ch
}

// scanWhitespace scans the next whitespace
func (s *CACMScanner) scanWhitespace() {
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !unicode.IsSpace(ch) {
			s.unread()
			break
		}
	}
}

// scanIdentifiant scans the next identifiant
// it returns "token" false if it's not actually an identifiant
func (s *CACMScanner) scanIdentifiant() (string, bool) {
	ch := s.read()
	tmp := s.peek()
	// we check it's really an identifiant, only one character and it's a letter
	if !unicode.IsSpace(tmp) || ch < 'A' || ch > 'Z' {
		s.unread()
		s.unread()
		return s.scanToken(), false
	}
	return "." + string(ch), true
}

func (s *CACMScanner) scanToken() string {
	// buffer to store the character
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == eof {
			break
		}
		if !tokenMember(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

// isCommonWord returns wether the word is part of the common word list
// It expects a sorted list, as provide by NewParser
func (s *CACMScanner) isCommonWord(lit string) bool {
	if len(lit) < 3 {
		return true
	}
	_, found := s.commonWord[lit]
	return found
}

func (s *CACMScanner) addToken(lit string) {
	// then the only token is the id
	if s.field == title {
		s.title.WriteString(lit)
	}
	// token are all token seen in document
	s.doc.addToken(lit)

	// cleaned and not common words are used for search
	if s.isCommonWord(lit) {
		return
	}
	s.doc.addWord(lit)
}

// Scan reads the next "word"
func (s *CACMScanner) Scan(c chan *Document) {
	for {
		ch := s.read()
		switch {
		case unicode.IsSpace(ch):
			s.unread()
			s.scanWhitespace()
			if s.field == title {
				s.title.WriteRune(' ')
			}
		case ch == '.':
			lit, isIdent := s.scanIdentifiant()
			if !isIdent {
			}
			s.field = identToField(lit)
			if s.field == id {
				if s.id != 0 {
					// Add the previous document
					s.doc.Title = s.title.String()
					// Send the document
					s.doc.calculFreqs()
					c <- s.doc
				}
				// Reset the document
				s.doc = newDocument()
				s.doc.Url = fmt.Sprintf("/cacm/%d", s.id)
				s.title.Reset()
				s.id++
			}
		case tokenMember(ch):
			s.unread()
			lit := s.scanToken()
			s.addToken(lit)
		case ch == eof:
			// Add the previous document
			s.doc.Title = s.title.String()
			// Send the document
			s.doc.calculFreqs()
			c <- s.doc
			close(c)
			return
		}
	}
}
