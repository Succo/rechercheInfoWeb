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

	fmt.Printf("Size of the vocabulary %d\n", parser.IndexSize())
	fmt.Printf("Number of token %d\n", parser.TokenSize())
}
