// Perf is structures listing information about the performances of the indexes
package main

import (
	"os"
	"time"
)

type Perf struct {
	Name string
	// Parsing is the time taken to parse all documents
	// build the temporary index and add metadata
	Parsing time.Duration
	// IDF is the time taken to build the IDf
	IDF time.Duration
	// Indexing is the time taken to build the trie
	Indexing time.Duration
	// Serialization is the time taken to serialize the whole Search struct
	// excluding the Perf obviously
	Serialization time.Duration
	// Total time taken
	TotalTime time.Duration
	// Index is the size of the docsID list
	Index uint64
	// Title the size of the list of titles
	Titles uint64
	// Total size of the indexes
	TotalSize uint64
	// Initial size of the corpus
	Initial uint64
	// Ratio between Total and initial
	Ratio float64
}

// newCACMPerf returns a perf object with the initial total size hardcoded
func newCACMPerf() Perf {
	return Perf{Initial: 2187734}
}

// newCS276Perf returns a perf object with the initial total size hardcoded
func newCS276Perf() Perf {
	return Perf{Initial: 429808000}
}

// getFinalValues complete the perf object
func (p Perf) getFinalValues() Perf {
	index, err := os.Lstat("indexes/" + p.Name + ".index")
	if err != nil {
		panic(err)
	}
	p.Index = uint64(index.Size())
	titles, err := os.Lstat("indexes/" + p.Name + ".titles")
	if err != nil {
		panic(err)
	}
	p.Titles = uint64(titles.Size())
	p.TotalSize = p.Index + p.Titles
	p.TotalTime = p.Parsing + p.IDF + p.Indexing + p.Serialization
	p.Ratio = float64(p.TotalSize) / float64(p.Initial)
	return p
}
