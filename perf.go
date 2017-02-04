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
	// Trie is the size of the prefix tree
	Trie uint64
	// Weight is the size of the the tfidf values
	Weight uint64
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
	trie, err := os.Lstat("indexes/" + p.Name + ".trie")
	if err != nil {
		panic(err)
	}
	p.Trie = uint64(trie.Size())
	weight, err := os.Lstat("indexes/" + p.Name + ".weight")
	if err != nil {
		panic(err)
	}
	p.Weight = uint64(weight.Size())
	p.TotalSize = p.Index + p.Trie + p.Weight
	p.TotalTime = p.Parsing + p.IDF + p.Indexing + p.Serialization
	p.Ratio = float64(p.TotalSize) / float64(p.Initial)
	return p
}
