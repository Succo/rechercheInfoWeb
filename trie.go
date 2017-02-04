// A trie implementation for the index
package main

import (
	"bytes"
	"encoding/gob"
	"os"
)

// Root is the root of the prefix tree
// Refs is a list of all references, nodes store pointer to it
type Root struct {
	Deltas []uint
	TfIdfs []float64
	Node   *Node
}

// Node implements a node of the tree
type Node struct {
	Sons  []*Node
	Radix [][]byte
	Start int
	End   int
}

func NewTrie(size int) *Root {
	// We fix the slice capacity to avoid further allocation
	deltas := make([]uint, 0, size)
	tfidfs := make([]float64, 0, size)
	return &Root{Node: &Node{}, Deltas: deltas, TfIdfs: tfidfs}
}

func emptyNode(start, end int) *Node {
	return &Node{Start: start, End: end}
}

func trieFromIndex(deltas map[string][]uint, tfidfs map[string][]float64, size int) *Root {
	r := NewTrie(size)
	for w, delta := range deltas {
		// That's because the first element is only a counter
		// Used when caluculating deltas
		r.set([]byte(w), delta[1:], tfidfs[w])
	}
	return r
}

func (r *Root) set(w []byte, deltas []uint, tfidfs []float64) {
	// calculate the start and end pointer for w
	start := len(r.Deltas)
	end := start + len(deltas)
	// Append the new deltas to the ref array
	r.Deltas = append(r.Deltas, deltas...)
	r.TfIdfs = append(r.TfIdfs, tfidfs...)

	// descends the tree to find the proper leaf
	cur := r.Node // node we are exploring
	shared := 0   // part of w already matched
	for {
		if shared == len(w) {
			cur.Start = start
			cur.End = end
			return
		}
		i := getMatchingNode(cur.Radix, w[shared])
		if i != -1 {
			rad := cur.Radix[i]
			// the two word share a prefix
			// calculate it's size
			size := longestPrefixSize(rad, w, shared)
			shared += size
			if size == len(rad) {
				cur = cur.Sons[i]
				continue
			}
			// split the vertice
			old := cur.Sons[i]
			new := &Node{
				Sons:  []*Node{old},
				Radix: [][]byte{rad[size:]},
			}
			// insert the new node in place
			cur.Radix[i] = rad[:size]
			cur.Sons[i] = new
			// keep iterating on the new node
			cur = new
		} else {
			// No son share a common prefix
			cur.Sons = append(cur.Sons, emptyNode(start, end))
			cur.Radix = append(cur.Radix, w[shared:])
			// bring the new node to it's place
			for j := len(cur.Radix) - 1; j > 0 &&
				bytes.Compare(cur.Radix[j-1], cur.Radix[j]) > 0; j-- {
				cur.Radix[j-1], cur.Radix[j] = cur.Radix[j], cur.Radix[j-1]
				cur.Sons[j-1], cur.Sons[j] = cur.Sons[j], cur.Sons[j-1]
			}
			break
		}
	}
}

// get returns the reference for a word
func (r *Root) get(w []byte) []Ref {
	cur := r.Node
	shared := 0
	for {
		if shared == len(w) {
			return r.buildRef(cur.Start, cur.End)
		}
		i := getMatchingNode(cur.Radix, w[shared])
		if i != -1 && bytes.HasPrefix(w[shared:], cur.Radix[i]) {
			shared += len(cur.Radix[i])
			cur = cur.Sons[i]
		} else {
			// No son share a common prefix
			return []Ref{}
		}
	}
}

// Serialize save to file the trie
func (r *Root) Serialize(name string) {
	index, err := os.Create("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	en := gob.NewEncoder(index)
	err = en.Encode(r.Deltas)
	if err != nil {
		panic(err)
	}
	index.Sync()
	index.Close()

	tfidf, err := os.Create("indexes/" + name + ".weight")
	if err != nil {
		panic(err)
	}
	defer tfidf.Close()
	err = Compress(r.TfIdfs, tfidf)
	if err != nil {
		panic(err)
	}
	tfidf.Sync()
	tfidf.Close()

	trie, err := os.Create("indexes/" + name + ".trie")
	if err != nil {
		panic(err)
	}
	defer trie.Close()
	en = gob.NewEncoder(trie)
	err = en.Encode(r.Node)
	if err != nil {
		panic(err)
	}
	trie.Sync()
	trie.Close()
}

// UnserializeTrie reloads the trie from files
func UnserializeTrie(name string) *Root {
	r := &Root{}
	index, err := os.Open("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	en := gob.NewDecoder(index)
	err = en.Decode(&r.Deltas)
	if err != nil {
		panic(err)
	}
	index.Close()

	tfidf, err := os.Open("indexes/" + name + ".weight")
	if err != nil {
		panic(err)
	}
	defer tfidf.Close()
	r.TfIdfs = UnCompress(tfidf)
	tfidf.Close()

	trie, err := os.Open("indexes/" + name + ".trie")
	if err != nil {
		panic(err)
	}
	defer trie.Close()
	en = gob.NewDecoder(trie)
	err = en.Decode(&r.Node)
	if err != nil {
		panic(err)
	}
	trie.Close()
	return r
}

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (r *Root) getInfIndex(maxID int) int {
	return r.Node.getInfIndex(maxID, r.Deltas)
}

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (n *Node) getInfIndex(maxID int, delta []uint) int {
	var indexSize int
	if n.Start-n.End != 0 && int(delta[n.Start]) <= maxID {
		indexSize++
	}
	for _, s := range n.Sons {
		indexSize += s.getInfIndex(maxID, delta)
	}
	return indexSize
}

// buildRed builds a Ref slice from a start and end position
func (r *Root) buildRef(start, end int) []Ref {
	deltas := r.Deltas[start:end]
	tfidfs := r.TfIdfs[start:end]
	var counter int
	refs := make([]Ref, len(deltas))
	for i, del := range deltas {
		counter += int(del)
		refs[i] = Ref{
			Id:    counter,
			TfIdf: tfidfs[i],
		}
	}
	return refs
}

// longestPrefixSize returns the longest prefix of rad and w
// with shared being the already matched part of w
// and assuming rad[0] == w[shared]
func longestPrefixSize(rad, w []byte, shared int) int {
	length := len(rad)
	if l := len(w) - shared; l < length {
		length = l
	}
	var i int
	for i = 1; i < length; i++ {
		if rad[i] != w[shared+i] {
			break
		}
	}
	return i
}

// getMatchingNode returns the index of the byte array that starts with a given byte
// or -1 if no match is found
func getMatchingNode(sons [][]byte, b byte) int {
	min := 0
	max := len(sons) - 1
	for min <= max {
		match := (max + min) / 2
		t := sons[match]
		if t[0] == b {
			return match
		} else if sons[match][0] < b {
			min = match + 1
		} else {
			max = match - 1
		}
	}
	return -1
}
