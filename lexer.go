package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

// cacmToUrl generates url from cacm id
func cacmToUrl(id int, title string) string {
	return fmt.Sprintf("/cacm/%d", id)
}

// Replacer is used to remplace _ by / in filename and get url
var replacer = strings.NewReplacer("_", "/")

// cs276ToUrl generates url from cs276 doc title (i.e file name)
func cs276ToUrl(id int, title string) string {
	return "https://" + replacer.Replace(title[2:]) // removes the [0-9]/ part of the title
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
	search.toUrl = cacmToUrl
	search.Perf = newCACMPerf()
	return buildSearchFromScanner(search, c)
}

// ParseCS276 creates a parser struct from the root folder of cs216 data
func ParseCS276(root string) *Search {
	cs276 := NewCS276Scanner(root)
	c := make(chan *Document)
	go cs276.Scan(c)
	search := emptySearch("cs276")
	search.toUrl = cs276ToUrl
	search.Perf = newCS276Perf()
	return buildSearchFromScanner(search, c)
}

func buildSearchFromScanner(search *Search, c chan *Document) *Search {
	now := time.Now()
	// Temporary storage for document id using delta
	deltas := make(map[string][]uint)
	tfidfs := make(map[string][]float64)

	// Number of document
	var count uint
	// Total number of document/word pairs
	var pairs int
	for doc := range c {
		search.AddDocMetaData(doc)
		for w, score := range doc.Scores {
			d, found := deltas[w]
			if !found {
				// The first element is actually a counter
				deltas[w] = []uint{count, count}
			} else {
				delta := count - d[0]
				d[0] = count
				deltas[w] = append(d, delta)
			}
			tfidfs[w] = append(tfidfs[w], score)
			pairs++
		}
		count++
	}
	search.Size = int(count)
	search.Perf.Parsing = time.Since(now)
	log.Printf("%s parsed in  %s \n", search.Corpus, time.Since(now).String())

	now = time.Now()
	// Now that all documents are known, we can fully calculate tf-idf
	calculateIDF(tfidfs, count)
	search.Perf.IDF = time.Since(now)
	log.Printf("%s IDF calculated in  %s \n", search.Corpus, time.Since(now).String())

	// Then we build the *real* index using a prefix tree
	now = time.Now()
	trie := trieFromIndex(deltas, tfidfs, pairs)
	search.Perf.Indexing = time.Since(now)
	log.Printf("%s index built in  %s \n", search.Corpus, time.Since(now).String())
	search.Index = trie
	search.Stat = getStat(search)
	search.Perf.Name = search.Corpus
	return search
}

func calculateIDF(tfidfs map[string][]float64, size uint) {
	factor := float64(size)
	for w, tfs := range tfidfs {
		idf := math.Log(factor * 1 / float64(len(tfs)))
		for i, tf := range tfs {
			// At this point tfidfs only contains Tf
			tfidfs[w][i] = tf * idf
		}
	}
}
