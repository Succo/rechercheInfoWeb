package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
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
	s      *Search
}

func NewPreCallCalculator(s *Search) *PreCallCalculator {
	return &PreCallCalculator{s: s}
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
			}
			qid++
		case bytes.HasPrefix(ln, []byte(".W")):
			inQBloc = true
		case inQBloc:
			buf.Write(ln)
		// Seems to suffice for all block indicator
		case bytes.HasPrefix(ln, []byte(".")):
			inQBloc = false
		}
	}
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
func (p *PreCallCalculator) Draw() {
	name := plotDir + "precision_recall" + plotFile
	if _, err := os.Stat(name); err == nil {
		// the file exist, whe assume it's the plot
		log.Println("Not generating precision recall as file already exist")
		return
	}

	now := time.Now()
	log.Println("Generating precision recall graph")
	plt, err := plot.New()
	if err != nil {
		panic(err)
	}

	plt.X.Label.Text = "Recall"
	plt.Y.Label.Text = "Precision"

	// Generate a semi-random color palette for graphs
	colors, err := colorful.HappyPalette(len(weightName))

	// Semaphore channel to wait for lines
	sem := make(chan bool)
	// iterate over all weight function in parrallel
	for i := range weightName {
		go func(i int) {
			wn := weightName[i]
			pts := make(plotter.XYs, 10)
			for j := 0; j < 10; j++ {
				pts[j].X, pts[j].Y = p.GetAvgRecallPrec((j*2)+1, i)
			}
			line, err := plotter.NewLine(pts)
			if err != nil {
				panic(err)
			}
			line.Color = colors[i]
			line.Width = vg.Points(2)
			plt.Add(line)
			plt.Legend.Add(wn, line)
			sem <- true
		}(i)
	}

	// Wait all graphs are generated
	for i := 0; i < len(weightName); i++ {
		<-sem
	}

	// Generate and save the graph
	if err = plt.Save(20*vg.Centimeter, 20*vg.Centimeter, name); err != nil {
		panic(err)
	}
	log.Printf("Precision recall graph generated in %s", time.Since(now).String())
}

// GetAvgPrecision get the Recall for n doc with weight function wf
func (p *PreCallCalculator) GetAvgRecallPrec(n, wf int) (recall, precision float64) {
	var refs []Ref
	var valid int
	for i, q := range p.queries {
		refs = VectorQuery(p.s, q, weight(wf))
		if len(refs) > n {
			refs = refs[:n]
		}
		valid = getNumberOfValidAns(refs, p.answer[i])
		recall += float64(valid) / float64(len(p.answer[i]))
		precision += float64(valid) / float64(n)
	}
	recall = recall / float64(len(p.queries))
	precision = precision / float64(len(p.queries))
	return recall, precision
}

func getNumberOfValidAns(refs []Ref, ans []int) int {
	var count int
	for _, ref := range refs {
		if len(ans) == 0 {
			break
		} else if ref.Id == ans[0] {
			count++
		} else if ref.Id > ans[0] {
			ans = ans[1:]
		}
	}
	return count
}
