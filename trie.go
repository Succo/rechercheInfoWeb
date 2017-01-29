// A trie implementation for the index
package main

import "bytes"

// Node implements a node of the tree
type Node struct {
	Refs  []Ref
	Sons  []*Node
	Radix [][]byte
}

func NewTrie() *Node {
	return &Node{}
}

func emptyNode(ref Ref) *Node {
	return &Node{Refs: []Ref{ref}}
}

func (n *Node) add(w []byte, ref Ref) {
	cur := n    // node we are exploring
	shared := 0 // part of w already matched
	for {
		if shared == len(w) {
			cur.Refs = append(cur.Refs, ref)
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
				Refs:  make([]Ref, 0),
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
			cur.Sons = append(cur.Sons, emptyNode(ref))
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

func (n *Node) get(w []byte) []Ref {
	cur := n
	shared := 0
	for {
		if shared == len(w) {
			return cur.Refs
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

// get InfIndex walks the tree
// returns the number of key wich are in a doc with index < maxID
func (n *Node) getInfIndex(maxID int) int {
	var indexSize int
	if len(n.Refs) != 0 && n.Refs[0].Id <= maxID {
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
