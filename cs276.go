package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	goroutineNumber = 20
)

// scanToken reads contiguous letter to read a word
func scanToken(s *bufio.Reader) string {
	// buffer to store the character
	var buf bytes.Buffer
	for {
		ch, _, err := s.ReadRune()
		if err != nil {
			break
		}
		if !tokenMember(ch) {
			s.UnreadRune()
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

// CS276Scanner will walk the buffer and return characters
type CS276Scanner struct {
	root   string
	dirs   []string
	toScan chan string
	wg     *sync.WaitGroup
}

// NewCS276Scanner create a CS276Scanner from a root dir string
func NewCS276Scanner(root string) *CS276Scanner {
	var wg sync.WaitGroup
	toScan := make(chan string, 100)
	return &CS276Scanner{
		root:   root,
		toScan: toScan,
		wg:     &wg,
	}
}

// scan processes string from the toScan channel
// it sends parsed document to the channel
func (s *CS276Scanner) scan(c chan *Document) {
	for filename := range s.toScan {
		doc := newDocument()
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
		scanner := bufio.NewReader(file)
		for {
			ch, _, err := scanner.ReadRune()
			if err != nil {
				break
			}
			if tokenMember(ch) {
				scanner.UnreadRune()
				w := scanToken(scanner)
				// all lexeme are compted as "seen"
				doc.addToken(w)
				doc.addWord(w)
			}
		}
		doc.calculTf()
		c <- doc
		file.Close()
	}
	s.wg.Done()
}

// Scan will send scanned doc to the channel using multiple goroutine to parse them
func (s *CS276Scanner) Scan(c chan *Document) {
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
	// goroutine parsing files
	for i := 0; i < goroutineNumber; i++ {
		go s.scan(c)
	}
	// the sync group should be the same size as the number of goroutines
	s.wg.Add(goroutineNumber)
	// goroutine to close the chan when all goroutines are done
	go func() {
		s.wg.Wait()
		close(c)
	}()
}
