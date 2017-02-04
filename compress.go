package main

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/golang/snappy"
)

func Compress(vals []float64, w io.Writer) error {
	snap := snappy.NewBufferedWriter(w)
	for _, v := range vals {
		_, err := snap.Write(Float64ToBytes(v))
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

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
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
		vals = append(vals, Float64FromBytes(buf))
	}
	return vals
}

func Float64FromBytes(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}
