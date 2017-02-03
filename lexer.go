package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
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
	return buildSearchFromScanner(search, c)
}

// ParseCS276 creates a parser struct from the root folder of cs216 data
func ParseCS276(root string) *Search {
	cs276 := NewCS276Scanner(root)
	c := make(chan *Document)
	go cs276.Scan(c)
	search := emptySearch("cs276")
	search.toUrl = cs276ToUrl
	return buildSearchFromScanner(search, c)
}

func buildSearchFromScanner(search *Search, c chan *Document) *Search {
	// Temporary storage for document id using delta
	deltas := make(map[string][]uint)
	tfidfs := make(map[string][]float64)

	// Number of document
	var count uint
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
		}
		count++
	}
	search.Size = int(count)
	// Now that all documents are known, we can fully calculate tf-idf
	calculateIDF(tfidfs, count)
	// Then we build the *real* index using a prefix tree
	trie := trieFromIndex(deltas, tfidfs)
	search.Index = trie
	search.Stat = getStat(search, "cs276")
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
