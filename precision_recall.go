package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
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

const (
	// The number of point used in the averaged graph
	approx int = 128
)

// PreCallCalculator is the struct that caluclate Precision and Recall from queries and answer
type PreCallCalculator struct {
	// Queries is a list of string corresponding to queries
	Queries []string
	// Answer is a list list fo valid doc ID for each query
	Answer [][]int
	// Valid is a list of valid graphs (i.e not empty)
	Valids []int
}

// Point is a value in a precision/recall graph
type Point struct {
	X, Y float64
}

// NewPreCallCalculator returns an empty PreCallCalculator
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
				p.Queries = append(p.Queries, buf.String())
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
	p.Queries = append(p.Queries, buf.String())
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
		length = len(p.Answer)
		if aid > length {
			//append a new empty slice
			p.Answer = append(p.Answer, []int{})
			length++
		}
		// Substract 1 from ans to compensate search starting from 0
		p.Answer[length-1] = append(p.Answer[length-1], ans-1)
	}
}

// Draw generates the precision/recall graph
func (p *PreCallCalculator) Draw(cacm *Search) {
	dir := path.Join(graphs, "precision_recall")
	if _, err := os.Stat(dir); err != nil {
		//the directory isn't there, generating one
		os.Mkdir(dir, 0777)
	}

	now := time.Now()
	log.Println("Generating precision recall graphs")
	// Semaphore channel to wait for all graph
	sem := make(chan bool)

	// Generate a semi-random color palette for graphs
	colors, err := colorful.HappyPalette(len(weightName))
	if err != nil {
		panic(err)
	}

	valids := make([]int, len(p.Queries))
	// small hack, valids[i] == i if the graph for query i is valid
	// array are initialised with 0, so valids[0] mist be != 0
	valids[0] = -1

	// Store the average for all valid queries
	var Average [total][approx]float64
	var Size [total]int

	for i := range p.Queries {
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
			for wf := 0; wf < total; wf++ {
				refs := VectorQuery(cacm, p.Queries[i], weight(wf))

				// Number of effectively valid answer
				var effective int
				valid := float64(len(p.Answer[i]))
				pts := make(plotter.XYs, 0)
				for count, ref := range refs {
					if contains(p.Answer[i], ref.Id) {
						effective++
						recall := float64(effective) / valid
						precision := float64(effective) / float64(count+1)
						pts = append(pts, Point{recall, precision})
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

				for index := 0; index < approx; index++ {
					Average[wf][index] += getApproxValue(pts, index)
				}
				Size[wf] += 1

			}
			if !useful {
				sem <- true
				return
			}
			valids[i] = i

			if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, file); err != nil {
				panic(err)
			}
			sem <- true
		}(i)
	}

	// Wait for all graphs to be generated
	for i := 0; i < len(p.Queries); i++ {
		<-sem
	}

	var count int
	newQ := make([]string, 0, len(p.Queries))
	for i, val := range valids {
		if i == val {
			count++
			p.Valids = append(p.Valids, val)
			newQ = append(newQ, p.Queries[i])
		}
	}
	p.Queries = newQ

	for j := 0; j < total; j++ {
		for x := 0; x < approx; x++ {
			Average[j][x] /= float64(count)
		}
	}
	// Graph for the averages
	file := path.Join(dir, "avg.svg")
	plt, err := plot.New()
	if err != nil {
		panic(err)
	}

	plt.X.Label.Text = "Recall"
	plt.Y.Label.Text = "Precision"
	step := 1 / float64(approx)
	var pos float64
	for wf := 0; wf < total; wf++ {
		pts := make(plotter.XYs, approx)
		pos = 0
		for i := 0; i < approx; i++ {
			pts[i].Y = Average[wf][i]
			pts[i].X = pos
			pos += step
		}
		line, err := plotter.NewLine(pts)
		if err != nil {
			panic(err)
		}
		line.Color = colors[wf]
		line.Width = vg.Points(2)
		plt.Add(line)
		wn := weightName[wf]
		plt.Legend.Add(wn, line)
	}
	if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, file); err != nil {
		panic(err)
	}

	// Generate and save the graph
	log.Printf("Precision recall graph generated in %s", time.Since(now).String())
}

// Serialize saves to file the data in PreCallCalculator
func (p *PreCallCalculator) Serialize() {
	precall, err := os.Create(path.Join("indexes", "cacm.precall"))
	if err != nil {
		panic(err)
	}
	defer precall.Close()
	en := gob.NewEncoder(precall)
	err = en.Encode(p)
	if err != nil {
		panic(err)
	}
}

// UnserializePreCallCalculator loads a serializes PeCallCalculator
func UnserializePreCallCalculator() *PreCallCalculator {
	var p *PreCallCalculator
	precall, err := os.Open(path.Join("indexes", "cacm.precall"))
	if err != nil {
		panic(err)
	}
	defer precall.Close()
	en := gob.NewDecoder(precall)
	err = en.Decode(&p)
	return p
}

func contains(haystack []int, needle int) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

// getApproxValue returns the approximate value in index/approx
// for the function having pts as its value
// It uses liners approximation, pts should have at least 2 points
func getApproxValue(pts plotter.XYs, index int) float64 {
	x := float64(index) / float64(approx)
	pos := -1
	// Get the last point smaller that x
	for i, pt := range pts {
		if pt.X > x {
			pos = i
			break
		}
	}
	// get two points for linear interpolation
	var p1, p2 Point
	if pos == -1 || pos == len(pts)-1 {
		// take the two last point as x > pt.x for all pt in pts
		p1 = pts[len(pts)-2]
		p2 = pts[len(pts)-1]
	} else {
		p1 = pts[pos]
		p2 = pts[pos+1]
	}
	a := (p1.Y - p2.Y) / (p1.X - p2.X)
	b := p1.Y - a*p1.X
	return a*x + b
}
