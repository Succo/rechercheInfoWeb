package main

import (
	"bufio"
	"bytes"
	"io"
)

// Character represent a lexical token
type Character int

const (
	Illegal Character = iota
	EOF
	WS

	Identifiant
	Token
)

var eof = rune(0)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	// Not precise enough, but should not be a problem here
	return !isWhitespace(ch) && ch != '.' && ch != eof
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) peek() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		ch = eof
	}
	s.r.UnreadRune()
	return ch
}

// scanWhitespace scans the next whitespace
func (s *Scanner) scanWhitespace() (ch Character, lit string) {
	// buffer to store contigous whitespace character
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}
	return WS, buf.String()
}

// scanIdentifiant scans the next identifiant, a two runes keyword
func (s *Scanner) scanIdentifiant() (Character, string) {
	ch := s.read()
	tmp := s.peek()
	// we check it's really an identifiant, only one character and it's a letter
	if !isWhitespace(tmp) || ch < 'A' || ch > 'Z' {
		s.unread()
		s.unread()
		return s.scanToken()
	}
	return Identifiant, "." + string(ch)
}

func (s *Scanner) scanToken() (Character, string) {
	// buffer to store contigous whitespace character
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	for {
		ch := s.read()
		if ch == eof {
			break
		}
		if !isLetter(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}
	return Token, buf.String()
}

// Scan reads the next "word"
func (s *Scanner) Scan() (Character, string) {
	ch := s.read()
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	}
	if ch == '.' {
		return s.scanIdentifiant()
	}
	if isLetter(ch) {
		s.unread()
		return s.scanToken()
	}
	if ch == eof {
		return EOF, ""
	}
	return Illegal, ""
}
