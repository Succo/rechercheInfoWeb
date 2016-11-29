package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

var cacmFile string
var commonWordFile string

func init() {
	flag.StringVar(&cacmFile, "cacm", "data/CACM/cacm.all", "Path to cacm file")
	flag.StringVar(&commonWordFile, "common_word", "data/CACM/common_words", "Path to common_word file")
}

func main() {
	flag.Parse()
	cacm, err := os.Open(cacmFile)
	if err != nil {
		panic(err)
	}
	defer cacm.Close()

	commonWord, err := os.Open(commonWordFile)
	if err != nil {
		panic(err)
	}
	defer commonWord.Close()

	var cw []string
	scanner := bufio.NewScanner(commonWord)
	for scanner.Scan() {
		cw = append(cw, scanner.Text())
	}

	parser := NewParser(cacm, cw)
	parser.Parse()

	corpusSize := parser.CorpusSize()
	fmt.Printf("For the whole corpus :\n")
	fmt.Printf("Size of the vocabulary %d\n", parser.IndexSize(corpusSize))
	fmt.Printf("Number of token %d\n", parser.TokenSize(corpusSize))
	fmt.Printf("For half the corpus :\n")
	fmt.Printf("Size of the vocabulary %d\n", parser.IndexSize(corpusSize/2))
	fmt.Printf("Number of token %d\n", parser.TokenSize(corpusSize/2))
}
