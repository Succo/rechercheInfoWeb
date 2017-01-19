package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var cacmFile string
var commonWordFile string
var plotFile string
var cs276File string
var cacmEnc string
var cs276Enc string

const (
	outputFormat = `For the whole %s corpus (%d documents) :
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
	flag.StringVar(&plotFile, "plot", "", "Common ending for plot file (extension can be different)")
	flag.StringVar(&cs276File, "cs276", "data/CS276/pa1-data", "Path to cs276 root folder")
	flag.StringVar(&cacmEnc, "serializedCacm", "", "File path to serialized index for cacm")
	flag.StringVar(&cs276Enc, "serializedCS276", "", "File path to serialized index for cs276")
}

//func main() {
//	flag.Parse()
//	var cacmSearch *Search
//	var cs276Search *Search
//	if cacmEnc == "" {
//		fmt.Println("Building cacm index from scratch")
//		cacm, err := os.Open(cacmFile)
//		if err != nil {
//			panic(err)
//		}
//		defer cacm.Close()
//		cacmParser := NewCACMParser(cacm, commonWordFile)
//		cacmSearch = cacmParser.Parse()
//		cacm.Close()
//	} else {
//		fmt.Println("Loading cacm index from file")
//		cacmSearch = NewSearch(cacmEnc)
//	}
//
//	printDetails(cacmSearch, "cacm")
//
//	if plotFile != "" {
//		draw(cacmSearch, "cacm")
//	}
//	if cacmEnc == "" {
//		cacmSearch.Serialize("cacm")
//	}
//
//	fmt.Println() // empty line
//	if cs276Enc == "" {
//		fmt.Println("Building cs276 index from scratch")
//		now := time.Now()
//		cs276Parser := NewCS276Parser(cs276File)
//		cs276Search = cs276Parser.Parse()
//		fmt.Printf("It took %s \n", time.Since(now).String())
//	} else {
//		fmt.Println("Loading cs276 index from file")
//		cs276Search = NewSearch(cs276Enc)
//	}
//	printDetails(cs276Search, "cs276")
//	if plotFile != "" {
//		draw(cs276Search, "cs276")
//	}
//	if cs276Enc == "" {
//		cs276Search.Serialize("cs276")
//	}
//	dynamicSearch(cacmSearch, cs276Search)
//}
func main() {
	serve()
}

func printDetails(search *Search, name string) {
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
