// Encoder.go stores function related to encding the index
package main

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/golang/snappy"
)

// Serialize save to file the trie
func (r *Root) Serialize(name string) {
	index, err := os.Create("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	snap := snappy.NewBufferedWriter(index)

	encodeInt(snap, r.count)
	r.Node.Encode(snap)

	err = snap.Close()
	if err != nil {
		panic(err.Error())
	}
	err = index.Close()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("done")
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

	r.count = decodeInt(snap)

	r.Node = &Node{}
	r.Node.Decode(snap)

	return r
}

// Encode writes to a Buffer data
// Used to implements GobEncode for Root
// Schema is
// len(Ref)
// [len(Ref)][total]float64 weights
// [len(Ref)]int ids
// len(sons)
// [len(sons] len(str) str
// [len(sons)] *Node
func (n *Node) Encode(encoder io.Writer) {
	encodeInt(encoder, len(n.Refs))
	for _, ref := range n.Refs {
		for i := 0; i < total; i++ {
			encodeFloat(encoder, ref.Weights[i])
		}
	}
	for _, ref := range n.Refs {
		encodeInt(encoder, ref.Id)
	}

	encodeStringSlice(encoder, n.Radix)
	for _, sons := range n.Sons {
		sons.Encode(encoder)
	}
}

// Decode decodes from an io.reader
// Used to implements GobEncode for Root
// Schema is
// len(Ref)
// [len(Ref)][total]float64 weights
// [len(Ref)]int ids
// len(sons)
// [len(sons] len(str) str
// [len(sons)] *Node
func (n *Node) Decode(decoder io.Reader) {
	length := decodeInt(decoder)
	fmt.Println(length)
	n.Refs = make([]Ref, length)
	for i := 0; i < length; i++ {
		for j := 0; j < total; j++ {
			n.Refs[i].Weights[j] = decodeFloat(decoder)
		}
	}
	for i := 0; i < int(length); i++ {
		n.Refs[i].Id = decodeInt(decoder)
	}

	n.Radix = decodeStringSlice(decoder)
	n.Sons = make([]*Node, len(n.Radix))
	for i := 0; i < len(n.Radix); i++ {
		n.Sons[i] = &Node{}
		n.Sons[i].Decode(decoder)
	}
}

// encodeInt writes an int to w
func encodeInt(w io.Writer, n int) {
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = uint8(n)
		n >>= 8
	}
	_, err := w.Write(buf)
	if err != nil {
		panic(err.Error())
	}
}

// decodeInt reads an int from r
func decodeInt(r io.Reader) int {
	buf := make([]byte, 8)
	_, err := r.Read(buf)
	if err != nil {
		panic(err.Error())
	}
	var n uint64
	for _, b := range buf {
		n = n<<8 | uint64(b)
	}
	return int(n)
}

// encodeFloat writes a float64 to w
// taken from https://github.com/golang/go/blob/964639cc338db650ccadeafb7424bc8ebb2c0f6c/src/encoding/gob/encode.go#L204
func encodeFloat(w io.Writer, f float64) {
	u := math.Float64bits(f)
	var v uint64
	for i := 0; i < 8; i++ {
		v <<= 8
		v |= u & 0xFF
		u >>= 8
	}
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = uint8(v)
		v >>= 8
	}
	_, err := w.Write(buf)
	if err != nil {
		panic(err.Error())
	}
}

// decodeFloat reads a float64 from r
func decodeFloat(r io.Reader) float64 {
	buf := make([]byte, 8)
	_, err := r.Read(buf)
	if err != nil {
		panic(err.Error())
	}
	var u uint64
	for _, b := range buf {
		u = u<<8 | uint64(b)
	}

	var v uint64
	for i := 0; i < 8; i++ {
		v <<= 8
		v |= u & 0xFF
		u >>= 8
	}
	return math.Float64frombits(v)
}

func encodeStringSlice(w io.Writer, str []string) {
	encodeInt(w, len(str))
	for _, rad := range str {
		b := []byte(rad)
		encodeInt(w, len(b))
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

func decodeStringSlice(r io.Reader) []string {
	length := decodeInt(r)

	str := make([]string, length)
	for i := 0; i < length; i++ {
		strlen := decodeInt(r)
		fmt.Println(length, strlen)
		rad := make([]byte, strlen)
		r.Read(rad)
		str[i] = string(rad)
		fmt.Println(str[i])
	}
	return str
}
