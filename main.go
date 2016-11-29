package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var cacmFile string
var commonWordFile string
var plotFile string

func init() {
	flag.StringVar(&cacmFile, "cacm", "data/CACM/cacm.all", "Path to cacm file")
	flag.StringVar(&commonWordFile, "common_word", "data/CACM/common_words", "Path to common_word file")
	flag.StringVar(&plotFile, "plot", "cacm_plot.png", "Path to output plot file")
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
	fmt.Printf("For the whole corpus (%d) :\n", corpusSize)
	fmt.Printf("Size of the vocabulary %d\n", parser.IndexSize(corpusSize))
	fmt.Printf("Number of token %d\n", parser.TokenSize(corpusSize))
	fmt.Printf("For half the corpus (%d):\n", corpusSize/2)
	fmt.Printf("Size of the vocabulary %d\n", parser.IndexSize(corpusSize/2))
	fmt.Printf("Number of token %d\n", parser.TokenSize(corpusSize/2))
	draw(parser)
}

func draw(parser *Parser) {
	corpusSize := parser.CorpusSize()

	plt, err := plot.New()
	if err != nil {
		panic(err)
	}

	plt.Title.Text = "Heaps law plot"
	plt.X.Label.Text = "Text size"
	plt.Y.Label.Text = "Distinct vocabulary"

	pts := make(plotter.XYs, corpusSize/10+1)
	for i := 0; i < corpusSize; i += 10 {
		pts[i/10].X = float64(parser.IndexSize(i))
		pts[i/10].Y = float64(parser.TokenSize(i))
	}

	err = plotutil.AddLines(plt, "cacm", pts)
	if err != nil {
		panic(err)
	}

	if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, plotFile); err != nil {
		panic(err)
	}
}
