package main

import "unicode"

var eof = rune(0)

func tokenMember(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '\'' || ch == '-' || ch == '/'
}

// Scanner is an interface reading words one at a time
// implemented by CACMscanner and CS276Scanner
type Scanner interface {
	Scan(c chan *Document)
}
