package main

import (
	"encoding/gob"
	"flag"
	"log"
	"os"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

var buildIndex bool

const (
	plotFile       = ".svg"
	cacmFile       = "data/CACM/cacm.all"
	commonWordFile = "data/CACM/common_words"
	cs276File      = "data/CS276/pa1-data"
)

func init() {
	flag.BoolVar(&buildIndex, "index", false, "-index=true to build index from scratch")
}

func main() {
	log.Println("RIW server started")
	flag.Parse()
	var cacm *Search
	var cs276 *Search
	// Get cacm and build related tools
	if buildIndex {
		log.Println("Building cacm index from scratch")
		source, err := os.Open(cacmFile)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		cacm = ParseCACM(source, commonWordFile)
		source.Close()
	} else {
		log.Println("Loading cacm index from file")
		gob.RegisterName("cacmR", cacmRetriever{})
		cacm = NewSearch("cacm")
	}

	if plotFile != "" {
		draw(cacm, "cacm")
	}
	if buildIndex {
		gob.RegisterName("cacmR", cacmRetriever{})
		cacm.Serialize("cacm")
	}

	// get cs276 and build related tools
	if buildIndex {
		log.Println("Building cs276 index from scratch")
		now := time.Now()
		cs276 = ParseCS276(cs276File)
		log.Printf("cs276 index built in  %s \n", time.Since(now).String())
	} else {
		log.Println("Loading cs276 index from file")
		gob.RegisterName("cs276R", cs276Retriever{})
		cs276 = NewSearch("cs276")
	}

	if plotFile != "" {
		draw(cs276, "cs276")
	}
	if buildIndex {
		gob.RegisterName("cs276R", cs276Retriever{})
		cs276.Serialize("cs276")
	}
	serve(cacm, cs276)
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
