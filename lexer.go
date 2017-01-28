package main

import (
	"bufio"
	"io"
	"os"

	porterstemmer "github.com/reiver/go-porterstemmer"
)

func cleanWord(word string) string {
	stem := porterstemmer.StemString(word)
	return stem
}

// ParseCACM creates a cacm scanner, a search struct and connects them
func ParseCACM(r io.Reader, commonWordFile string) *Search {
	commonWord, err := os.Open(commonWordFile)
	if err != nil {
		panic(err)
	}
	defer commonWord.Close()

	cw := make(map[string]bool)
	scanner := bufio.NewScanner(commonWord)
	for scanner.Scan() {
		cw[scanner.Text()] = true
	}

	cacm := NewCACMScanner(r, cw)

	c := make(chan *Document)
	go cacm.Scan(c)
	search := emptySearch("cacm")
	ids := make([]int64, 0)
	for doc := range c {
		search.AddDocument(doc)
		ids = append(ids, doc.pos)
	}
	search.Retriever = &cacmRetriever{Ids: ids}
	search.Stat = getStat(search, "cacm")
	return search
}

// ParseCS276 creates a parser struct from the root folder of cs216 data
func ParseCS276(root string) *Search {
	cs276 := NewCS276Scanner(root)
	c := make(chan *Document)
	go cs276.Scan(c)
	search := emptySearch("cs276")
	for doc := range c {
		search.AddDocument(doc)
	}
	search.Retriever = &cs276Retriever{}
	search.Stat = getStat(search, "cs276")
	return search
}
