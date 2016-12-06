package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var cacmFile string
var commonWordFile string
var plotFile string
var cs276 string

const (
	outputFormat = `For the whole %s corpus (%d documents) :
It took %s
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
	flag.StringVar(&plotFile, "plot", "_plot.svg", "Path to output plot file")
	flag.StringVar(&cs276, "cs276", "data/CS276/pa1-data", "Path to cs 276 root folder")
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

	now := time.Now()
	cacmParser := NewCACMParser(cacm, cw)
	cacmSearch := cacmParser.Parse()

	printDetails(cacmSearch, "cacm", time.Since(now))
	draw(cacmSearch, "cacm")

	fmt.Println() // empty line
	now = time.Now()
	cs276Parser := NewCS276Parser(cs276)
	cs276Search := cs276Parser.Parse()

	printDetails(cs276Search, "cs276", time.Since(now))
	draw(cs276Search, "cs276")
}

func printDetails(search *Search, name string, calculTime time.Duration) {
	corpusSize := search.CorpusSize()
	tokenSize := search.TokenSize(corpusSize)
	halfTokenSize := search.TokenSize(corpusSize / 2)
	vocabSize := search.IndexSize(corpusSize)
	halfVocabSize := search.IndexSize(corpusSize / 2)

	// Heaps law calculation
	b := (math.Log(float64(tokenSize)) - math.Log(float64(halfTokenSize))) / (math.Log(float64(vocabSize)) - math.Log(float64(halfVocabSize)))
	k := float64(tokenSize) / (math.Pow(float64(vocabSize), b))

	fmt.Printf(
		outputFormat,
		name,
		corpusSize,
		calculTime.String(),
		vocabSize,
		tokenSize,
		corpusSize/2,
		halfVocabSize,
		halfTokenSize,
		b,
		k,
		k*math.Pow(1000000000, b))
}

func draw(search *Search, name string) {
	corpusSize := search.CorpusSize()

	plt, err := plot.New()
	if err != nil {
		panic(err)
	}

	plt.Title.Text = "Heaps law plot for " + name
	plt.X.Label.Text = "Text size"
	plt.Y.Label.Text = "Distinct vocabulary"

	pts := make(plotter.XYs, 100)
	for i := 0; i < 100; i++ {
		pts[i].X = float64(search.IndexSize(i * corpusSize / 100))
		pts[i].Y = float64(search.TokenSize(i * corpusSize / 100))
	}

	err = plotutil.AddLines(plt, name, pts)
	if err != nil {
		panic(err)
	}

	if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, name+plotFile); err != nil {
		panic(err)
	}
}
