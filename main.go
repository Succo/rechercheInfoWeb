package main

import (
	"bufio"
	"flag"
	"image/color"
	"log"
	"os"
	"path"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"

	"net/http"
	_ "net/http/pprof"
)

var buildIndex, buildPrecall bool

const (
	graphs         = "graphs"
	cacmFile       = "data/CACM/cacm.all"
	commonWordFile = "data/CACM/common_words"
	cs276File      = "data/CS276/pa1-data"
)

func init() {
	flag.BoolVar(&buildIndex, "index", false, "-index to build index from scratch")
	flag.BoolVar(&buildPrecall, "precall", false, "-precall to rebuild precision/recall data")
}

func main() {
	log.Println("Starting riw server")
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	flag.Parse()
	c := make(chan *Search)
	// Build a set of common words
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

	go buildCACM(c, cw)
	go buildCS276(c, cw)
	var cacm, cs276 *Search
	var precall *PreCallCalculator
	var s *Search
	for i := 0; i < 2; i++ {
		s = <-c
		if s.Corpus == "cacm" {
			cacm = s
			if buildPrecall {
				precall = NewPreCallCalculator()
				precall.Populate("data/CACM/query.text", "data/CACM/qrels.text")
				precall.Draw(cacm)
				precall.Serialize()
			} else {
				precall = UnserializePreCallCalculator()
			}
		} else {
			cs276 = s
		}
	}
	serve(cacm, cs276, precall)
}

// draw generates heaps law graph
func draw(search *Search) {
	name := search.Corpus
	file := path.Join(graphs, name+".svg")
	if _, err := os.Stat(file); err == nil {
		// the file exist, whe assume it's the plot
		return
	}

	corpusSize := search.Size

	plt, err := plot.New()
	if err != nil {
		panic(err)
	}

	plt.X.Label.Text = "Index size"
	plt.Y.Label.Text = "Distinct vocabulary"

	pts := make(plotter.XYs, 100)
	for i := 0; i < 100; i++ {
		pts[i].X = float64(search.TokenSize(i * corpusSize / 100))
		pts[i].Y = float64(search.IndexSize(i * corpusSize / 100))
	}
	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	line.Color = color.RGBA{R: 10, G: 174, B: 194, A: 255}
	line.Width = vg.Points(2)

	plt.Add(line)
	plt.Legend.Add(name, line)

	if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, file); err != nil {
		panic(err)
	}
}

func buildCACM(c chan *Search, cw map[string]bool) {
	var cacm *Search
	if buildIndex {
		log.Println("Building cacm index from scratch")
		source, err := os.Open(cacmFile)
		if err != nil {
			panic(err)
		}
		defer source.Close()
		cacm = ParseCACM(source, cw)
		source.Close()
		draw(cacm)
		cacm.Serialize()
	} else {
		log.Println("Loading cacm index from file")
		cacm = UnserializeSearch("cacm")
		cacm.toUrl = cacmToUrl
	}
	c <- cacm
}

func buildCS276(c chan *Search, cw map[string]bool) {
	var cs276 *Search
	if buildIndex {
		log.Println("Building cs276 index from scratch")
		cs276 = ParseCS276(cs276File, cw)
		draw(cs276)
		cs276.Serialize()
	} else {
		log.Println("Loading cs276 index from file")
		cs276 = UnserializeSearch("cs276")
		cs276.toUrl = cs276ToUrl
	}
	c <- cs276
}
