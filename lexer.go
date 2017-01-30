package main

import (
	"bufio"
	"io"
	"math"
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
	// Store doc ID to position in cacm.all pointer
	ids := make([]int64, 0)
	// Temporary storage for document
	index := make(map[string][]Ref)
	// Number of document
	var count int
	for doc := range c {
		search.AddDocMetaData(doc)
		for w, score := range doc.Scores {
			index[w] = append(index[w], Ref{count, score})
		}
		ids = append(ids, doc.pos)
		count++
	}
	search.Retriever = &cacmRetriever{Ids: ids}
	calculateIDF(index, count)
	search.Stat = getStat(search, "cacm")
	return search
}

// ParseCS276 creates a parser struct from the root folder of cs216 data
func ParseCS276(root string) *Search {
	cs276 := NewCS276Scanner(root)
	c := make(chan *Document)
	go cs276.Scan(c)
	search := emptySearch("cs276")
	// Store doc ID to position in cacm.all pointer
	ids := make([]int64, 0)
	// Temporary storage for document
	index := make(map[string][]Ref)
	// Number of document
	var count int
	for doc := range c {
		search.AddDocMetaData(doc)
		for w, score := range doc.Scores {
			index[w] = append(index[w], Ref{count, score})
		}
		ids = append(ids, doc.pos)
		count++
	}
	search.Retriever = &cs276Retriever{}
	calculateIDF(index, count)
	search.Stat = getStat(search, "cs276")
	return search
}

func calculateIDF(index map[string][]Ref, size int) {
	factor := float64(size)
	for w, refs := range index {
		wordFactor := math.Log(factor * 1 / float64(len(refs)))
		for i, ref := range refs {
			ref.TfIdf = ref.TfIdf * wordFactor
			index[w][i] = ref
		}
	}
}
