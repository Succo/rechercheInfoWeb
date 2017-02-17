package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	goroutineNumber = 20
)

// CS276Scanner will walk the buffer and return characters
type CS276Scanner struct {
	root   string
	dirs   []string
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
func (s *CS276Scanner) scan(c chan *Document, sem chan bool) {
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
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			w := scanner.Text()
			// all lexeme are compted as "seen"
			doc.addToken(w)
			doc.addWord(w)
		}
		doc.calculScore()
		s.trie.addDoc(doc)
		c <- doc
		file.Close()
	}
	sem <- true
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
