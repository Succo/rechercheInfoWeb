// This package provides wrapperfor snappy to compress the weight array
package main

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/golang/snappy"
)

func Compress(vals []weights, w io.Writer) error {
	snap := snappy.NewBufferedWriter(w)
	var bits uint64
	var err error
	// 8 byte for a flaot64
	buf := make([]byte, 8)
	for _, v := range vals {
		bits = math.Float64bits(v[raw])
		binary.BigEndian.PutUint64(buf, bits)
		_, err = snap.Write(buf)
		if err != nil {
			return err
		}
		bits = math.Float64bits(v[norm])
		binary.BigEndian.PutUint64(buf, bits)
		_, err = snap.Write(buf)
		if err != nil {
			return err
		}
		bits = math.Float64bits(v[half])
		binary.BigEndian.PutUint64(buf, bits)
		_, err = snap.Write(buf)
		if err != nil {
			return err
		}
	}
	err = snap.Close()
	if err != nil {
		return err
	}
	return nil
}

func UnCompress(r io.Reader) []weights {
	snap := snappy.NewReader(r)
	// 8 bytes for a float64
	buf := make([]byte, 8)
	vals := make([]weights, 0)
	var val weights
	var err error
	var bits uint64
	for {
		_, err = snap.Read(buf)
		if err != nil {
			break
		}
		bits = binary.BigEndian.Uint64(buf)
		val[raw] = math.Float64frombits(bits)
		_, err = snap.Read(buf)
		if err != nil {
			break
		}
		bits = binary.BigEndian.Uint64(buf)
		val[norm] = math.Float64frombits(bits)
		_, err = snap.Read(buf)
		if err != nil {
			break
		}
		bits = binary.BigEndian.Uint64(buf)
		val[half] = math.Float64frombits(bits)
		vals = append(vals, val)
	}
	return vals
}
