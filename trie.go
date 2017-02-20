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
	count int
}

// Node implements a node of the tree
type Node struct {
	// rw is a RWMutex, can be hold by either
	// 1 writer or many reader
	rw sync.RWMutex
	// Sons and Radix holds information
	Sons  []*Node
	Radix []string
	// Refs hold information about the word ending at this node
	Refs []Ref
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
func (r *Root) add(w string, id int, tfidf weights) {
	// descends the tree to find the proper leaf
	cur := r.Node             // node we are exploring
	var shared, i, length int // shared: part of w already matche,
	rad := ""                 // buffer for radix
	ref := Ref{id, tfidf}
	for {
		if shared == len(w) {
			cur.rw.Lock()
			cur.Refs = append(cur.Refs, ref)
			for j := len(cur.Refs) - 1; j > 0 &&
				cur.Refs[j-1].Id > cur.Refs[j].Id; j-- {
				cur.Refs[j-1], cur.Refs[j] = cur.Refs[j], cur.Refs[j-1]
			}
			cur.rw.Unlock()
			return
		}
	MainInsert:
		cur.rw.RLock()
		i = getMatchingNode(cur.Radix, w[shared])
		if i != len(cur.Radix) && cur.Radix[i][0] == w[shared] {
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
			if len(cur.Radix) != length {
				// the node has been updated, restart reading
				cur.rw.Unlock()
				goto MainInsert
			}
			// No son share a common prefix
			new := &Node{
				Refs: []Ref{ref},
			}
			cur.Sons = append(cur.Sons, new)
			cur.Radix = append(cur.Radix, "")
			copy(cur.Sons[i+1:], cur.Sons[i:])
			copy(cur.Radix[i+1:], cur.Radix[i:])
			cur.Sons[i] = new
			cur.Radix[i] = w[shared:]
			// bring the new node to it's place
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
			refs := r.buildRef(cur.Refs)
			cur.rw.RUnlock()
			return refs
		}
		i := getMatchingNode(cur.Radix, w[shared])
		if i != len(cur.Radix) && strings.HasPrefix(w[shared:], cur.Radix[i]) {
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

// buildRed builds a Ref slice it's needed to make sure we don't update in place
// the initial slice
func (r *Root) buildRef(in []Ref) []Ref {
	out := make([]Ref, len(in))
	copy(out, in)
	return out
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
	idf := math.Log(factor / float64(len(n.Refs)))
	for i, ref := range n.Refs {
		n.Refs[i].Weights = scale(ref.Weights, idf)
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
	return r.Node.getInfIndex(maxID)
}

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (n *Node) getInfIndex(maxID int) int {
	var indexSize int
	if len(n.Refs) > 0 && n.Refs[0].Id < maxID {
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

// getMatchingNode returns the index of the
// first string that starts with a byte >= b
// or len(sons) if no match is found
func getMatchingNode(sons []string, b byte) int {
	if len(sons) < 10 {
		for i, son := range sons {
			if son[0] >= b {
				return i
			}
		}
		return len(sons)
	}

	// If the slice is long use binary search
	// better tuning needed here
	min := 0
	max := len(sons) - 1
	var match int
	for min < max {
		match = min + (max-min)/2
		if sons[match][0] < b {
			min = match + 1
		} else {
			max = match
		}
	}
	return min
}
