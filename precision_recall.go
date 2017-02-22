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

// PreCallCalculator is the struct that caluclate Precision and Recall from queries and answer
type PreCallCalculator struct {
	// Queries is a list of string corresponding to queries
	Queries []string
	// Answer is a list list fo valid doc ID for each query
	Answer [][]int
	// Valid is a list of valid graphs (i.e not empty)
	Valids []int
	// MAP is the Mean Average Precision Value
	MAP [total]float64
	// descirption of the differentts weight function
	Descrpt [total]string
}

// Point is a value in a precision/recall graph
type Point struct {
	X, Y float64
}

// NewPreCallCalculator returns an empty PreCallCalculator
func NewPreCallCalculator() *PreCallCalculator {
	return &PreCallCalculator{Descrpt: weightName}
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
	// array are initialised with 0, so valids[0] must be != 0
	valids[0] = -1

	// Store all plots used,
	// It is used to get averages
	var Average [total][]*plotter.Function
	for wf := 0; wf < total; wf++ {
		Average[wf] = make([]*plotter.Function, len(p.Queries))
	}

	for i := range p.Queries {
		go func(i int) {
			file := path.Join(dir, strconv.Itoa(i)+".svg")
			plt := getPlot()

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
						// Adds a point with X = recall, Y = precision
						pts = append(pts,
							Point{float64(effective) / valid,
								float64(effective) / float64(count+1)})
					}
				}
				if len(pts) < 2 {
					// No valid point, most likely "invalid query", like on the author
					continue
				}

				f := funcFromPoints(pts)
				useful = true
				f.Color = colors[wf]
				plt.Add(f)
				wn := weightName[wf]
				plt.Legend.Add(wn, f)

				Average[wf][i] = f
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

	// Graph for the averages
	file := path.Join(dir, "avg.svg")
	plt := getPlot()

	var f [total]*plotter.Function
	for wf := 0; wf < total; wf++ {
		f[wf] = averageFunction(Average[wf])
		f[wf].Color = colors[wf]
		plt.Add(f[wf])
		wn := weightName[wf]
		plt.Legend.Add(wn, f[wf])
		p.MAP[wf] = getMAP(f[wf])
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

func getPlot() *plot.Plot {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.X.Label.Text = "Recall"
	p.Y.Label.Text = "Precision"
	p.X.Min = 0.0
	p.X.Max = 1.0
	p.Y.Min = 0.0
	p.Y.Max = 1.0
	p.Legend.Top = true
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

// funcFromPoints generate a function for the plotter interface
// the function respects precision recall graph logic
func funcFromPoints(pts plotter.XYs) *plotter.Function {
	f := plotter.NewFunction(func(x float64) float64 {
		var max float64
		for _, pt := range pts {
			if pt.X > x && pt.Y > max {
				max = pt.Y
			}
		}
		return max
	})
	f.Width = vg.Points(2)
	f.Samples = 256
	return f
}

// averageFunction returns a function for the plotter interface
// the function is the average of all non nil function taken in argument
func averageFunction(funcs []*plotter.Function) *plotter.Function {
	var nonNil int
	for _, f := range funcs {
		if f != nil {
			nonNil++
		}
	}
	f := plotter.NewFunction(func(x float64) float64 {
		var avg float64
		for _, f := range funcs {
			if f != nil {
				avg += f.F(x)
			}
		}
		return avg / float64(nonNil)
	})
	f.Width = vg.Points(2)
	f.Samples = 256
	return f
}

func getMAP(f *plotter.Function) float64 {
	sample := 256
	var avg float64
	for i := 0; i < sample; i++ {
		avg += f.F(float64(i) / float64(sample))
	}
	return avg / float64(sample)
}
