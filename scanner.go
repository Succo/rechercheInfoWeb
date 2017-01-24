package main

import "unicode"

// Character represent a lexical token
type Character int

const (
	// Illegal are the ignored charater
	Illegal Character = iota
	// EOF is a special character for the End Of File
	EOF
	// WS is the character for all whitespaces
	WS

	// Identifiant is for the marqueurs
	Identifiant
	// Token is all the non special words
	Token
)

var eof = rune(0)

func tokenMember(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '\'' || ch == '-' || ch == '/'
}

// Scanner is an interface reading words one at a time
// implemented by CACMscanner and CS276Scanner
type Scanner interface {
	Scan(c chan *Document)
}
