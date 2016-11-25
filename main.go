package main

import (
	"bufio"
	"os"
)

func main() {
	cacm, err := os.Open("data/CACM/cacm.all")
	if err != nil {
		panic(err)
	}
	defer cacm.Close()

	common_word, err := os.Open("data/CACM/common_words")
	if err != nil {
		panic(err)
	}
	defer common_word.Close()

	var cw []string
	scanner := bufio.NewScanner(common_word)
	for scanner.Scan() {
		cw = append(cw, scanner.Text())
	}

	parser := NewParser(cacm, cw)
	parser.Parse()
}
