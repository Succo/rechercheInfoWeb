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
	// Index is the size of the docsID list
	Index uint64
	// Trie is the size of the prefix tree
	Trie uint64
	// Weight is the size of the the tfidf values
	Weight uint64
}

func (p Perf) addSerializedSizes() Perf {
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
	return p
}
