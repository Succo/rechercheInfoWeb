package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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

// a token is a scanned word as returned by worker
type token struct {
	word string
	ch   Character
}

// CS276Scanner will walk the buffer and return characters
type CS276Scanner struct {
	root      string
	dirs      []string
	toScan    chan string
	scanned   chan []token
	inProcess []token
	wg        sync.WaitGroup
}

// NewCS276Scanner create a CS276Scanner from a root dir string
func NewCS276Scanner(root string) *CS276Scanner {
	var wg sync.WaitGroup
	toScan := make(chan string, 100)
	s := &CS276Scanner{root: root, dirs: make([]string, 0), toScan: toScan, scanned: make(chan []token, 10), inProcess: make([]token, 0), wg: wg}
	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}
	// goroutine that will add all file to parse by reading dir in order
	go func() {
		for _, dir := range dirs {
			files, err := ioutil.ReadDir(s.root + "/" + dir.Name())
			if err != nil {
				fmt.Println(err)
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
		go s.scan()
	}
	// the sync group should be the same size as the number of goroutines
	s.wg.Add(goroutineNumber)
	// goroutine to close the chan when all goroutines are done
	go func() {
		s.wg.Wait()
		close(s.scanned)
	}()
	return s
}

// scan processes string from the toScan channel
// either adding new file to scan or :w
func (s *CS276Scanner) scan() {
	for filename := range s.toScan {
		// This whole section is hack to fit cacm data model
		// TODO update scanner design
		scanned := []token{
			token{word: ".I", ch: Identifiant},
			token{word: "", ch: Token},
			token{word: ".T", ch: Identifiant},
			token{word: filename, ch: Token},
			token{word: ".W", ch: Identifiant},
		}
		file, err := os.Open(s.root + "/" + filename)
		defer file.Close()
		if err != nil {
			fmt.Println(err)
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
				scanned = append(scanned, token{word: scanToken(scanner), ch: Token})
			}
		}
		s.scanned <- scanned
		file.Close()
	}
	s.wg.Done()
}

// Scan reads the next "word"
// Parsing folders as needed
func (s *CS276Scanner) Scan() (Character, string) {
	if len(s.inProcess) == 0 {
		toProcess, more := <-s.scanned
		if more {
			s.inProcess = toProcess
		} else {
			return EOF, ""
		}
	}
	token := s.inProcess[0]
	s.inProcess = s.inProcess[1:]
	return token.ch, token.word
}
