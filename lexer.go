package main

import (
	"fmt"
	"io"
	"log"
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
func ParseCACM(r io.Reader, cw map[string]bool) *Search {
	// index stored in a prefix trie
	trie := NewTrie()
	cacm := NewCACMScanner(r, cw, trie)

	c := make(chan *Document)
	go cacm.Scan(c)

	search := emptySearch("cacm", cw)
	search.toUrl = cacmToUrl
	search.Perf = newCACMPerf()
	search.Index = trie
	return buildSearchFromScanner(search, c)
}

// ParseCS276 creates a parser struct from the root folder of cs216 data
func ParseCS276(root string, cw map[string]bool) *Search {
	// index stored in a prefix trie
	trie := NewTrie()
	cs276 := NewCS276Scanner(root, trie)
	// chan for processed documents
	// metadata are handled in the main thread
	c := make(chan *Document, 100)
	go cs276.Scan(c)

	search := emptySearch("cs276", cw)
	search.toUrl = cs276ToUrl
	search.Perf = newCS276Perf()
	search.Index = trie
	return buildSearchFromScanner(search, c)
}

func buildSearchFromScanner(search *Search, c chan *Document) *Search {
	now := time.Now()

	// The main loop get parsed documents and deals with metadata
	for doc := range c {
		search.AddDocMetaData(doc)
	}
	search.Size = len(search.Tokens)
	// potentially, the index is not finished so time is innacurate
	// the mutex protects from incorrect read though
	search.Perf.Parsing = time.Since(now)
	log.Printf("%s parsed in  %s \n", search.Corpus, time.Since(now).String())

	now = time.Now()
	// Now that all documents are known, we can fully calculate tf-idf
	search.Index.calculateIDF(search.Size)
	search.Perf.IDF = time.Since(now)
	log.Printf("%s IDF calculated in  %s \n", search.Corpus, time.Since(now).String())

	log.Printf("%s index average sons count for non leaf node %f\n",
		search.Corpus,
		search.Index.getAverageSonsCount())

	search.Stat = getStat(search)
	search.Perf.Name = search.Corpus
	return search
}
