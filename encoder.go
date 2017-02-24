// Encoder.go stores function related to encding the index
package main

import (
	"bufio"
	"io"
	"log"
	"math"
	"os"
	"time"
)

// Serialize save to file the trie
func (r *Root) Serialize(name string) {
	now := time.Now()
	index, err := os.Create("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	buffered := bufio.NewWriter(index)

	buf := make([]byte, 8)
	encodeInt(buffered, r.count, buf)
	r.Node.Encode(buffered, buf)
	buffered.Flush()

	err = index.Close()
	if err != nil {
		panic(err.Error())
	}
	log.Printf("%s index serialization took %s", name, time.Since(now))
}

// UnserializeTrie reloads the trie from files
func UnserializeTrie(name string) *Root {
	now := time.Now()
	r := &Root{}
	index, err := os.Open("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	buffered := bufio.NewReader(index)

	buf := make([]byte, 8)
	r.count = decodeInt(buffered, buf)

	r.Node = &Node{}
	r.Node.Decode(buffered, buf)

	err = index.Close()
	if err != nil {
		panic(err.Error())
	}
	log.Printf("%s index unserialization took %s", name, time.Since(now))
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
func (n *Node) Encode(encoder io.Writer, buf []byte) {
	encodeInt(encoder, len(n.Refs), buf)
	for _, ref := range n.Refs {
		for i := 0; i < total; i++ {
			encodeFloat(encoder, ref.Weights[i], buf)
		}
	}
	for _, ref := range n.Refs {
		encodeInt(encoder, ref.Id, buf)
	}

	encodeStringSlice(encoder, n.Radix, buf)
	for _, sons := range n.Sons {
		sons.Encode(encoder, buf)
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
func (n *Node) Decode(decoder io.Reader, buf []byte) {
	length := decodeInt(decoder, buf)
	n.Refs = make([]Ref, length)
	for i := 0; i < length; i++ {
		for j := 0; j < total; j++ {
			n.Refs[i].Weights[j] = decodeFloat(decoder, buf)
		}
	}
	for i := 0; i < int(length); i++ {
		n.Refs[i].Id = decodeInt(decoder, buf)
	}

	n.Radix = decodeStringSlice(decoder, buf)
	n.Sons = make([]*Node, len(n.Radix))
	for i := 0; i < len(n.Radix); i++ {
		n.Sons[i] = &Node{}
		n.Sons[i].Decode(decoder, buf)
	}
}

// encodeInt writes an int to w
func encodeInt(w io.Writer, n int, buf []byte) {
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
func decodeInt(r io.Reader, buf []byte) int {
	read(buf, r)
	var n uint64
	for _, b := range buf {
		n = n<<8 | uint64(b)
	}
	return int(n)
}

// encodeFloat writes a float64 to w
// taken from https://github.com/golang/go/blob/964639cc338db650ccadeafb7424bc8ebb2c0f6c/src/encoding/gob/encode.go#L204
func encodeFloat(w io.Writer, f float64, buf []byte) {
	u := math.Float64bits(f)
	var v uint64
	for i := 0; i < 8; i++ {
		v <<= 8
		v |= u & 0xFF
		u >>= 8
	}
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
func decodeFloat(r io.Reader, buf []byte) float64 {
	read(buf, r)
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

func encodeStringSlice(w io.Writer, str []string, buf []byte) {
	encodeInt(w, len(str), buf)
	for _, rad := range str {
		b := []byte(rad)
		encodeInt(w, len(b), buf)
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

func decodeStringSlice(r io.Reader, buf []byte) []string {
	length := decodeInt(r, buf)

	str := make([]string, length)
	rad := make([]byte, 8)
	for i := 0; i < length; i++ {
		strlen := decodeInt(r, buf)
		if strlen > len(rad) {
			rad = make([]byte, strlen)
		}
		read(rad[:strlen], r)
		str[i] = string(rad[:strlen])
	}
	return str
}

// read wraps r.Read, making sure all reads are complete
func read(buf []byte, r io.Reader) {
	var read int
	length := len(buf)
	for read < length {
		n, err := r.Read(buf[read:length])
		if err != nil {
			panic(err.Error())
		}
		read += n
	}
}
