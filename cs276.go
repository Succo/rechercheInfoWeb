// CS276 scanner is the struct that will parse CS276 documents
// It does so concurrently using multiples goroutine to read document
// A first goroutine list all files available and send filename though a channel
// multiples worker (goroutineNumber) read this chan and process documents when available
// Processed documents are indexed concurrently and sent through a chan for metadata (titles...)
// This chan also serve to see when processing is finished (by closing it)
package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

const (
	goroutineNumber = 20
)

// CS276Scanner will walk the buffer and return characters
type CS276Scanner struct {
	root   string
	toScan chan string
	trie   *Root
}

// NewCS276Scanner create a CS276Scanner from a root dir string
func NewCS276Scanner(root string, trie *Root) *CS276Scanner {
	toScan := make(chan string, 100)
	return &CS276Scanner{
		root:   root,
		toScan: toScan,
		trie:   trie,
	}
}

// scan processes string from the toScan channel
// it sends parsed document to the channel
func (s *CS276Scanner) scan(c chan metadata, sem chan bool) {
	doc := newDocument()
	for filename := range s.toScan {
		doc.Title = filename
		// words of the title are added too
		words := strings.Split(filename, "_")
		for _, w := range words[1:] {
			doc.addToken(w)
			doc.addWord(w)
		}

		file, err := os.Open(s.root + "/" + filename)
		defer file.Close()
		if err != nil {
			log.Println(err)
			break
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(scanWords)
		for scanner.Scan() {
			w := BytesToString(scanner.Bytes())
			// all lexeme are compted as "seen"
			doc.addToken(w)
			doc.addWord(w)
		}
		s.trie.addDoc(doc)
		c <- metadataFromDoc(doc)
		doc.reset()
		file.Close()
	}
	sem <- true
}

// Scan will send scanned doc to the channel using multiple goroutine to parse them
func (s *CS276Scanner) Scan(c chan metadata) {
	dirs, err := ioutil.ReadDir(s.root)
	if err != nil {
		panic(err)
	}
	// goroutine that will add all file to parse by reading dir in order
	go func() {
		for _, dir := range dirs {
			files, err := ioutil.ReadDir(s.root + "/" + dir.Name())
			if err != nil {
				log.Println(err)
				continue
			}
			for _, file := range files {
				s.toScan <- (dir.Name() + "/" + file.Name())
			}
		}
		close(s.toScan)
	}()
	// Semaphore to wait for all routine to be done
	sem := make(chan bool, 2)
	// goroutine parsing files
	for i := 0; i < goroutineNumber; i++ {
		go s.scan(c, sem)
	}
	// goroutine to close the chan when all goroutines are done
	go func() {
		for i := 0; i < goroutineNumber; i++ {
			<-sem
		}
		close(c)
	}()
}

// BytesToString is an unsafe converter from []byte to string
func BytesToString(b []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{bytesHeader.Data, bytesHeader.Len}
	return *(*string)(unsafe.Pointer(&strHeader))
}

// scanWords is a split function for a Scanner that returns each
// space-separated word of text, with surrounding spaces deleted. It will
// never return an empty string. The definition of space is ' '
// it expects bytes and no special character
// should really only be used with CS276, And even then should not be used
// Addapted from https://golang.org/src/bufio/scan.go?s=12782:12860#L374
func scanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for ; start < len(data); start++ {
		if data[start] != ' ' {
			break
		}
	}
	// Scan until space, marking end of word.
	for i := start; i < len(data); i++ {
		if data[i] == ' ' {
			return i + 1, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	// Without the last character (newline)
	if atEOF && len(data) > start {
		return len(data), data[start : len(data)-1], nil
	}
	// Request more data.
	return start, nil, nil
}
