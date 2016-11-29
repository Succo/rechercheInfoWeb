package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var cacmFile string
var commonWordFile string
var plotFile string

const (
	outputFormat = `For the whole corpus (%d documents) :
Size of the vocabulary %d
Number of token %d
For half the corpus (%d documents):
Size of the vocabulary %d
Number of token %d

Heaps Law gives us:
b = %f
k = %f

For 1 million token we get %f as vocabulary size
`
)

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

	printDetails(parser)
	draw(parser)
}

func printDetails(parser *Parser) {
	corpusSize := parser.CorpusSize()
	tokenSize := parser.TokenSize(corpusSize)
	halfTokenSize := parser.TokenSize(corpusSize / 2)
	vocabSize := parser.IndexSize(corpusSize)
	halfVocabSize := parser.IndexSize(corpusSize / 2)

	// Heaps law calculation
	b := (math.Log(float64(tokenSize)) - math.Log(float64(halfTokenSize))) / (math.Log(float64(vocabSize)) - math.Log(float64(halfVocabSize)))
	k := float64(tokenSize) / (math.Pow(float64(vocabSize), b))

	fmt.Printf(
		outputFormat,
		corpusSize,
		vocabSize,
		tokenSize,
		corpusSize/2,
		halfVocabSize,
		halfTokenSize,
		b,
		k,
		k*math.Pow(1000000000, b))
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
