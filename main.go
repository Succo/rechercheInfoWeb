package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var cacmFile string
var commonWordFile string
var cs276File string
var cacmEnc string
var cs276Enc string

const (
	plotFile = ".svg"
)

func init() {
	flag.StringVar(&cacmFile, "cacm", "data/CACM/cacm.all", "Path to cacm file")
	flag.StringVar(&commonWordFile, "common_word", "data/CACM/common_words", "Path to common_word file")
	flag.StringVar(&cs276File, "cs276", "data/CS276/pa1-data", "Path to cs276 root folder")
	flag.StringVar(&cacmEnc, "serializedCacm", "", "File path to serialized index for cacm")
	flag.StringVar(&cs276Enc, "serializedCS276", "", "File path to serialized index for cs276")
}

func main() {
	log.Println("RIW server started")
	flag.Parse()
	var cacmSearch *Search
	var cs276Search *Search
	if cacmEnc == "" {
		log.Println("Building cacm index from scratch")
		cacm, err := os.Open(cacmFile)
		if err != nil {
			panic(err)
		}
		defer cacm.Close()
		cacmParser := NewCACMParser(cacm, commonWordFile)
		cacmSearch = cacmParser.Parse()
		cacm.Close()
	} else {
		log.Println("Loading cacm index from file")
		cacmSearch = NewSearch(cacmEnc)
	}

	if plotFile != "" {
		draw(cacmSearch, "cacm")
	}
	if cacmEnc == "" {
		cacmSearch.Serialize("cacm")
	}

	if cs276Enc == "" {
		log.Println("Building cs276 index from scratch")
		now := time.Now()
		cs276Parser := NewCS276Parser(cs276File)
		cs276Search = cs276Parser.Parse()
		log.Printf("cs276 index built in  %s \n", time.Since(now).String())
	} else {
		log.Println("Loading cs276 index from file")
		cs276Search = NewSearch(cs276Enc)
	}

	if plotFile != "" {
		draw(cs276Search, "cs276")
	}
	if cs276Enc == "" {
		cs276Search.Serialize("cs276")
	}
	serve(cacmSearch, cs276Search)
}

func draw(search *Search, name string) {
	if _, err := os.Stat(name + plotFile); err == nil {
		// the file exist, whe assume it's the plot
		return
	}

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
