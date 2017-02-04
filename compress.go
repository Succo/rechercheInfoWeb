// This package provides wrapperfor snappy to compress the weight array
package main

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/golang/snappy"
)

func Compress(vals []float64, w io.Writer) error {
	snap := snappy.NewBufferedWriter(w)
	// 8 byte for a flaot64
	buf := make([]byte, 8)
	for _, v := range vals {
		bits := math.Float64bits(v)
		binary.BigEndian.PutUint64(buf, bits)
		_, err := snap.Write(buf)
		if err != nil {
			return err
		}
	}
	err := snap.Close()
	if err != nil {
		return err
	}
	return nil
}

func UnCompress(r io.Reader) []float64 {
	snap := snappy.NewReader(r)
	// 8 bytes for a float64
	buf := make([]byte, 8)
	vals := make([]float64, 0)
	for {
		_, err := snap.Read(buf)
		if err != nil {
			break
		}
		bits := binary.BigEndian.Uint64(buf)
		vals = append(vals, math.Float64frombits(bits))
	}
	return vals
}
