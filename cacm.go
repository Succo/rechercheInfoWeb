package main

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

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
func (s *CACMScanner) scanWhitespace() token {
	// buffer to store contigous whitespace character
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !unicode.IsSpace(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}
	return token{buf.String(), WS}
}

// scanIdentifiant scans the next identifiant, a two runes keyword
func (s *CACMScanner) scanIdentifiant() token {
	ch := s.read()
	tmp := s.peek()
	// we check it's really an identifiant, only one character and it's a letter
	if !unicode.IsSpace(tmp) || ch < 'A' || ch > 'Z' {
		s.unread()
		s.unread()
		return s.scanToken()
	}
	return token{"." + string(ch), Identifiant}
}

func (s *CACMScanner) scanToken() token {
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
	return token{buf.String(), Token}
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
	// we store the lowest ID where the word was seen
	// it's easy since id are seen in order
	lit = cleanWord(lit)
	if s.isCommonWord(lit) {
		return
	}
	// We add non common word to the token list
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
			t := s.scanIdentifiant()
			if t.ch == Token {
				s.addToken(t.word)
				break
			}
			s.field = identToField(t.word)
			if s.field == id {
				if s.id != 0 {
					// Add the previous document
					s.doc.Title = s.title.String()
					// Send the document
					c <- s.doc
				}
				// Reset the document
				s.doc = newDocument()
				s.title.Reset()
				s.id++
			}
		case tokenMember(ch):
			s.unread()
			t := s.scanToken()
			s.addToken(t.word)
		case ch == eof:
			// Add the previous document
			s.doc.Title = s.title.String()
			// Send the document
			c <- s.doc
			close(c)
			return
		}
	}
}
