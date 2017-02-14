package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// PreCallCalculator is the struct that caluclate Precision and Recall from queries and answer
type PreCallCalculator struct {
	// queries is a list of string corresponding to queries
	queries []string
	// answer is a list list fo valid doc ID for each query
	answer [][]int
	// list of valid graphs (i.e not empty)
	valids []int
}

func NewPreCallCalculator() *PreCallCalculator {
	return &PreCallCalculator{}
}

// Populate adds all the needed values to tue queries and answer list
func (p *PreCallCalculator) Populate(query string, answer string) {
	// Counter when population the array
	var qid, aid int
	// Indicates when we are in a query bloc
	var inQBloc bool
	Q, err := os.Open(query)
	if err != nil {
		panic(err.Error())
	}
	scanner := bufio.NewScanner(Q)
	var buf bytes.Buffer
	var ln []byte
	for scanner.Scan() {
		ln = scanner.Bytes()
		switch {
		case bytes.HasPrefix(ln, []byte(".I")):
			if qid > 0 {
				p.queries = append(p.queries, buf.String())
				buf.Reset()
			}
			qid++
		case bytes.HasPrefix(ln, []byte(".W")):
			inQBloc = true
		// Seems to suffice for all block indicator
		case bytes.HasPrefix(ln, []byte(".")):
			inQBloc = false
		case inQBloc:
			buf.Write(ln)
			buf.Write([]byte(" "))
		}
	}
	p.queries = append(p.queries, buf.String())
	A, err := os.Open(answer)
	if err != nil {
		panic(err.Error())
	}
	scanner = bufio.NewScanner(A)
	var line string
	var ans, length int
	for scanner.Scan() {
		line = scanner.Text()
		fmt.Sscanf(line, "%d %d", &aid, &ans)
		length = len(p.answer)
		if aid > length {
			//append a new empty slice
			p.answer = append(p.answer, []int{})
			length++
		}
		// Substract 1 from ans to compensate search starting from 0
		p.answer[length-1] = append(p.answer[length-1], ans-1)
	}
}

// Draw generates the precision/recall graph
func (p *PreCallCalculator) Draw(cacm *Search) {
	dir := path.Join(graphs, "precision_recall")
	if _, err := os.Stat(dir); err == nil {
		// the file exist, whe assume it's the plot
		log.Println("Not generating precision recall as directory already exist")
		return
	}
	os.Mkdir(dir, 0777)

	now := time.Now()
	log.Println("Generating precision recall graphs")
	// Semaphore channel to wait for all graph
	sem := make(chan bool)

	// Generate a semi-random color palette for graphs
	colors, err := colorful.HappyPalette(len(weightName))
	if err != nil {
		panic(err)
	}

	valids := make([]int, len(p.queries))
	// small hack, valids[i] == i if the graph for query i is valid
	// array are initialised with 0, so valids[0] mist be != 0
	valids[0] = -1

	for i := range p.queries {
		go func(i int) {
			file := path.Join(dir, strconv.Itoa(i)+".svg")
			plt, err := plot.New()
			if err != nil {
				panic(err)
			}

			plt.X.Label.Text = "Recall"
			plt.Y.Label.Text = "Precision"

			// A boolean to check that line are added to the plot
			// don't draw uselate plot
			var useful bool
			// iterate over all weight function in parrallel
			for wf := range weightName {
				refs := VectorQuery(cacm, p.queries[i], weight(wf))

				// Number of effectively valid answer
				var effective int
				valid := float64(len(p.answer[i]))
				pts := make(plotter.XYs, 0)
				for count, ref := range refs {
					if contains(p.answer[i], ref.Id) {
						effective++
						recall := float64(effective) / valid
						precision := float64(effective) / float64(count+1)
						pts = append(pts, struct{ X, Y float64 }{recall, precision})
					}
				}
				if len(pts) < 2 {
					// No valid point, most likely "invalid query", like on the author
					continue
				}

				line, err := plotter.NewLine(pts)
				if err != nil {
					panic(err)
				}

				useful = true
				line.Color = colors[wf]
				line.Width = vg.Points(2)
				plt.Add(line)
				wn := weightName[wf]
				plt.Legend.Add(wn, line)
			}
			if !useful {
				sem <- true
				return
			}
			if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, file); err != nil {
				panic(err)
			}
			valids[i] = i
			sem <- true
		}(i)
	}

	// Wait for all graphs to be generated
	for i := 0; i < len(p.queries); i++ {
		<-sem
	}
	for i, val := range valids {
		if i == val {
			p.valids = append(p.valids, val)
		}
	}

	// Generate and save the graph
	log.Printf("Precision recall graph generated in %s", time.Since(now).String())
}

func contains(haystack []int, needle int) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}
