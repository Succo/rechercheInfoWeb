// A trie implementation for the index
// It supports concurrent access using mutex on all nodes
// So it can be built concurrently by all thread
// RWMutex should have a low impact once the index is built
package main

import (
	"encoding/gob"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/golang/snappy"
)

// Root is the root of the prefix tree
type Root struct {
	Node *Node
	// The Root of the tree counts the total number of documents
	// the count is protected by a lock
	mu    sync.Mutex
	count uint
}

// Node implements a node of the tree
type Node struct {
	// rw is a RWMutex, can be hold by either
	// 1 writer or many reader
	rw sync.RWMutex
	// Sons and Radix holds information
	Sons  []*Node
	Radix []string
	// Deltas and TfIdfs holds information about the word ending at this node
	Deltas []uint
	TfIdfs []weights
}

func NewTrie() *Root {
	return &Root{Node: &Node{}}
}

// addDoc adds all a document references to the trie
// It also generates the document ID
func (r *Root) addDoc(doc *Document) {
	r.mu.Lock()
	doc.Id = r.count
	r.count++
	r.mu.Unlock()
	for w, score := range doc.Scores {
		r.add(string(w), doc.Id, score)
	}
}

// add the weights and id to w
func (r *Root) add(w string, id uint, tfidf weights) {
	// descends the tree to find the proper leaf
	cur := r.Node             // node we are exploring
	var shared, i, length int // shared: part of w already matche,
	rad := ""                 // buffer for radix
	for {
		if shared == len(w) {
			cur.rw.Lock()
			cur.Deltas = append(cur.Deltas, id)
			cur.TfIdfs = append(cur.TfIdfs, tfidf)
			for j := len(cur.Deltas) - 1; j > 0 &&
				cur.Deltas[j-1] > cur.Deltas[j]; j-- {
				cur.Deltas[j-1], cur.Deltas[j] = cur.Deltas[j], cur.Deltas[j-1]
				cur.TfIdfs[j-1], cur.TfIdfs[j] = cur.TfIdfs[j], cur.TfIdfs[j-1]
			}
			cur.rw.Unlock()
			return
		}
	MainInsert:
		cur.rw.RLock()
		i = getMatchingNode(cur.Radix, w[shared])
		if i != -1 {
			rad = cur.Radix[i]
			// if cur.Radix is a complete prefix go down the trie
			if strings.HasPrefix(w[shared:], rad) {
				shared += len(rad)
				new := cur.Sons[i]
				cur.rw.RUnlock()
				cur = new
				continue
			}
			// Unlock reads, lock write
			cur.rw.RUnlock()
			cur.rw.Lock()
			if rad != cur.Radix[i] {
				// the node has been updated, restart reading
				cur.rw.Unlock()
				goto MainInsert
			}
			// the two word share a prefix
			// calculate it's size
			size := longestPrefixSize(rad, w, shared)
			shared += size
			// split the vertice
			old := cur.Sons[i]
			new := &Node{
				Sons:  []*Node{old},
				Radix: []string{rad[size:]},
			}
			// insert the new node in place
			cur.Radix[i] = rad[:size]
			cur.Sons[i] = new
			// keep iterating on the new node
			cur.rw.Unlock()
			cur = new
		} else {
			// Unlock reads, lock write
			length = len(cur.Radix)
			cur.rw.RUnlock()
			cur.rw.Lock()
			// Assert the node hasn't been updated inbetween
			if len(cur.Radix) != -1 {
				// the node has been updated, restart reading
				cur.rw.Unlock()
				goto MainInsert
			}
			// No son share a common prefix
			new := &Node{
				Deltas: []uint{id},
				TfIdfs: []weights{tfidf},
			}
			cur.Sons = append(cur.Sons, new)
			cur.Radix = append(cur.Radix, w[shared:])
			// bring the new node to it's place
			for j := len(cur.Radix) - 1; j > 0 &&
				strings.Compare(cur.Radix[j-1], cur.Radix[j]) > 0; j-- {
				cur.Radix[j-1], cur.Radix[j] = cur.Radix[j], cur.Radix[j-1]
				cur.Sons[j-1], cur.Sons[j] = cur.Sons[j], cur.Sons[j-1]
			}
			cur.rw.Unlock()
			break
		}
	}
}

// get returns the reference for a word
func (r *Root) get(w string) []Ref {
	cur := r.Node
	shared := 0
	for {
		cur.rw.RLock()
		if shared == len(w) {
			refs := r.buildRef(cur.Deltas, cur.TfIdfs)
			cur.rw.RUnlock()
			return refs
		}
		i := getMatchingNode(cur.Radix, w[shared])
		if i != -1 && strings.HasPrefix(w[shared:], cur.Radix[i]) {
			shared += len(cur.Radix[i])
			new := cur.Sons[i]
			cur.rw.RUnlock()
			cur = new
		} else {
			// No son share a common prefix
			return []Ref{}
		}
	}
}

// buildRed builds a Ref slice from a start and end position
func (r *Root) buildRef(deltas []uint, tfidfs []weights) []Ref {
	refs := make([]Ref, len(deltas))
	for i, del := range deltas {
		refs[i] = Ref{
			Id:      int(del),
			Weights: tfidfs[i],
		}
	}
	return refs
}

// calculateIDF calculateIDF in a concurrent maner
func (r *Root) calculateIDF(size uint) {
	factor := float64(size)
	for _, son := range r.Node.Sons {
		go son.calculateIDF(factor)
	}
}

// calculateIDF walks through the tree calculating IDF for all nodes
func (n *Node) calculateIDF(factor float64) {
	n.rw.RLock()
	idf := math.Log(factor / float64(len(n.TfIdfs)))
	for i, tf := range n.TfIdfs {
		n.TfIdfs[i] = scale(tf, idf)
	}
	for _, son := range n.Sons {
		son.calculateIDF(factor)
	}
	n.rw.RUnlock()
}

// Serialize save to file the trie
func (r *Root) Serialize(name string) {
	index, err := os.Create("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	snap := snappy.NewBufferedWriter(index)
	en := gob.NewEncoder(snap)
	err = en.Encode(r.Node)
	if err != nil {
		panic(err)
	}
	err = snap.Close()
	if err != nil {
		panic(err.Error())
	}
	index.Close()
}

// UnserializeTrie reloads the trie from files
func UnserializeTrie(name string) *Root {
	r := &Root{}
	index, err := os.Open("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	defer index.Close()
	snap := snappy.NewReader(index)
	en := gob.NewDecoder(snap)
	err = en.Decode(&r.Node)
	if err != nil {
		panic(err)
	}
	index.Close()
	return r
}

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (r *Root) getInfIndex(maxID int) int {
	return r.Node.getInfIndex(uint(maxID))
}

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (n *Node) getInfIndex(maxID uint) int {
	var indexSize int
	if len(n.Deltas) > 0 && n.Deltas[0] < maxID {
		indexSize++
	}
	for _, s := range n.Sons {
		indexSize += s.getInfIndex(maxID)
	}
	return indexSize
}

// longestPrefixSize returns the longest prefix of rad and w
// with shared being the already matched part of w
// and assuming rad[0] == w[shared]
func longestPrefixSize(rad, w string, shared int) int {
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

// getMatchingNode returns the index of the string that starts with a given byte
// or -1 if no match is found
func getMatchingNode(sons []string, b byte) int {
	min := 0
	max := len(sons) - 1
	var match int
	for min <= max {
		match = (max + min) / 2
		if sons[match][0] == b {
			return match
		} else if sons[match][0] < b {
			min = match + 1
		} else {
			max = match - 1
		}
	}
	return -1
}
