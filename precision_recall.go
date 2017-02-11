package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

// PreCallCalculator is the struct that caluclate Precision and Recall from queries and answer
type PreCallCalculator struct {
	// queries is a list of string corresponding to queries
	queries []string
	// answer is a list list fo valid doc ID for each query
	answer [][]int
}

func NewQuestion() *PreCallCalculator {
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
		case bytes.HasSuffix(ln, []byte(".I")):
			if qid > 0 {
				p.queries = append(p.queries, buf.String())
			}
			qid++
		case bytes.HasSuffix(ln, []byte(".W")):
			inQBloc = true
		case inQBloc:
			buf.Write(ln)
		// Seems to suffice for all block indicator
		case bytes.HasSuffix(ln, []byte(".")):
			inQBloc = false
		}
	}
	A, err := os.Open(query)
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
		}
		// Substract 1 from ans to compensate search starting from 0
		p.answer[length-1] = append(p.answer[length], ans-1)
	}
}
