package main

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// CACMCACMScanner will walk the buffer and return character
type CACMScanner struct {
	r *bufio.Reader
}

// NewCACMScanner create a CACMScanner from an io reader
func NewCACMScanner(r io.Reader) *CACMScanner {
	return &CACMScanner{bufio.NewReader(r)}
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
func (s *CACMScanner) scanWhitespace() (ch Character, lit string) {
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
	return WS, buf.String()
}

// scanIdentifiant scans the next identifiant, a two runes keyword
func (s *CACMScanner) scanIdentifiant() (Character, string) {
	ch := s.read()
	tmp := s.peek()
	// we check it's really an identifiant, only one character and it's a letter
	if !unicode.IsSpace(tmp) || ch < 'A' || ch > 'Z' {
		s.unread()
		s.unread()
		return s.scanToken()
	}
	return Identifiant, "." + string(ch)
}

func (s *CACMScanner) scanToken() (Character, string) {
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
	return Token, buf.String()
}

// Scan reads the next "word"
func (s *CACMScanner) Scan() (Character, string) {
	ch := s.read()
	if unicode.IsSpace(ch) {
		s.unread()
		return s.scanWhitespace()
	}
	if ch == '.' {
		return s.scanIdentifiant()
	}
	if tokenMember(ch) {
		s.unread()
		return s.scanToken()
	}
	if ch == eof {
		return EOF, ""
	}
	return Illegal, ""
}
