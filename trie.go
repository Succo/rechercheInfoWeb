// A trie implementation for the index
package main

// Node implements a node of the tree
type Node struct {
	refs  []*Ref
	sons  []*Node
	radix []string
}

func NewTrie() *Node {
	return &Node{}
}

func emptyNode(ref *Ref) *Node {
	return &Node{refs: []*Ref{ref}}
}

func (n *Node) add(w string, ref *Ref) {
	cur := n    // node we are exploring
	shared := 0 // part of w already matched
	for {
		if shared == len(w)-1 {
			cur.refs = append(cur.refs, ref)
			return
		}
		for i, rad := range cur.radix {
			if rad[0] != w[shared] {
				continue
			}
			// the two word share a prefix
			// calculate it's size
			size := 1
			for size < len(rad) && size < len(w)-shared &&
				rad[:size+1] == w[shared:size+1] {
				size++
			}
			// split the vertice
			new := &Node{
				refs:  cur.refs,
				sons:  cur.sons,
				radix: make([]string, len(cur.radix)),
			}
			for j := range new.radix {
				new.radix[j] = cur.radix[j][size:]
			}
			// insert the new node in place
			cur.radix[i] = rad[:size]
			cur.sons[i] = new
			// keep iterating on the new node
			cur = new
			shared += size
			continue
		}
		// No son share a common prefix
		cur.sons = append(cur.sons, emptyNode(ref))
		cur.radix = []string{w[shared:]}
		// bring the new node to it's place
		for j := len(cur.radix) - 1; j > 0 && cur.radix[j-1] > cur.radix[j]; j-- {
			cur.radix[j-1], cur.radix[j] = cur.radix[j], cur.radix[j-1]
			cur.sons[j-1], cur.sons[j] = cur.sons[j], cur.sons[j-1]
		}
		break
	}
}

func (n *Node) get(w string) []*Ref {
	cur := n
	shared := 0
	for {
		if shared == len(w) {
			return cur.refs
		}
		for i, rad := range cur.radix {
			if rad[0] != w[shared] {
				continue
			}
			// the two word share a prefix
			// calculate it's size
			size := 1
			for size < len(rad) && size < len(w)-shared {
				if rad[:size+1] == w[shared:size+1] {
					size++
				}
			}
			cur = cur.sons[i]
			shared += size
		}
		// No son share a common prefix
		return []*Ref{}
	}
}
