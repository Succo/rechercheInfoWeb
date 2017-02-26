// Encoder.go stores function related to encding the index
// other classes are written to files using gob a encoding library inclued in go
// besause the library needs to "read" the underlying type of the data it was
// way too slow to use directly on a recursive data structure like a trie

package main

import (
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/golang/snappy"
)

const (
	uint64Size = 8
)

// Serialize save to file the trie
func (r *Root) Serialize(name string) {
	now := time.Now()
	index, err := os.Create("indexes/" + name + ".index")
	if err != nil {
		panic(err)
	}
	buffered := snappy.NewBufferedWriter(index)

	// 9 is the size of a uint64 + 1, see encodeUint for details
	buf := make([]byte, 9)
	encodeUInt(buffered, uint(r.count), buf)
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
	buffered := snappy.NewReader(index)

	buf := make([]byte, 9)
	r.count = int(decodeUInt(buffered, buf))

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
// [len(Ref)]int ids // delta encoded
// len(sons)
// [len(sons] len(str) str
// [len(sons)] *Node
func (n *Node) Encode(encoder io.Writer, buf []byte) {
	encodeUInt(encoder, uint(len(n.Refs)), buf)
	if len(n.Refs) > 0 {
		for _, ref := range n.Refs {
			for i := 0; i < total; i++ {
				encodeFloat(encoder, ref.Weights[i], buf)
			}
		}
		encodeUInt(encoder, uint(n.Refs[0].Id), buf)
		for i := 1; i < len(n.Refs); i++ {
			// delta encoding
			encodeUInt(encoder, uint(n.Refs[i].Id-n.Refs[i-1].Id), buf)
		}
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
// [len(Ref)]int ids // delta encoded
// len(sons)
// [len(sons] len(str) str
// [len(sons)] *Node
func (n *Node) Decode(decoder io.Reader, buf []byte) {
	length := int(decodeUInt(decoder, buf))
	n.Refs = make([]Ref, length)
	for i := 0; i < length; i++ {
		for j := 0; j < total; j++ {
			n.Refs[i].Weights[j] = decodeFloat(decoder, buf)
		}
	}
	if length > 0 {
		n.Refs[0].Id = int(decodeUInt(decoder, buf))
		for i := 1; i < length; i++ {
			n.Refs[i].Id = int(decodeUInt(decoder, buf)) + n.Refs[i-1].Id
		}
	}

	n.Radix = decodeStringSlice(decoder, buf)
	n.Sons = make([]*Node, len(n.Radix))
	for i := 0; i < len(n.Radix); i++ {
		n.Sons[i] = &Node{}
		n.Sons[i].Decode(decoder, buf)
	}
}

// encodeUInt writes an uint to w
// if the int is smaller than 128 it's written as is
// otherwise the number of bytes is written followed by the values
func encodeUInt(w io.Writer, n uint, buf []byte) {
	if n <= 0x7F {
		w.Write([]byte{uint8(n)})
		return
	}
	i := uint64Size
	for n > 0 {
		buf[i] = uint8(n)
		n >>= 8
		i--
	}
	buf[i] = uint8(i - 8)
	_, err := w.Write(buf[i : uint64Size+1])
	if err != nil {
		panic(err.Error())
	}
}

// decodeUInt reads an uint from r
// if the int is smaller than 128 then it's the first byte
// otherwise the first byte is the number of byte
func decodeUInt(r io.Reader, buf []byte) uint {
	read(buf[:1], r)
	if buf[0] <= 0x7F {
		return uint(buf[0])
	}
	i := -int(int8(buf[0]))
	if i > uint64Size {
		panic("Error encoded interger too big")
	}
	read(buf[:i], r)
	var n uint64
	for _, b := range buf[:i] {
		n = n<<8 | uint64(b)
	}
	return uint(n)
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
	encodeUInt(w, uint(v), buf)
}

// decodeFloat reads a float64 from r
func decodeFloat(r io.Reader, buf []byte) float64 {
	u := uint64(decodeUInt(r, buf))
	var v uint64
	for i := 0; i < 8; i++ {
		v <<= 8
		v |= u & 0xFF
		u >>= 8
	}
	return math.Float64frombits(v)
}

// encodeStringSlice encodes a slice of string
// first encoding the length of the slice
// then the len of each string followed by the bytes composing the string
func encodeStringSlice(w io.Writer, str []string, buf []byte) {
	encodeUInt(w, uint(len(str)), buf)
	for _, rad := range str {
		b := []byte(rad)
		encodeUInt(w, uint(len(b)), buf)
		_, err := w.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

// decodeStringSlice decodes a slice of string
// first decoding the length of the slice
// then the len of each string followed by the bytes composing the string
func decodeStringSlice(r io.Reader, buf []byte) []string {
	length := int(decodeUInt(r, buf))
	str := make([]string, length)
	// Uses this as a potentially bigger buffer
	rad := make([]byte, 8)
	for i := 0; i < length; i++ {
		strlen := int(decodeUInt(r, buf))
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
