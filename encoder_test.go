package main

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestEncodeInt(t *testing.T) {
	var buf bytes.Buffer
	temp := make([]byte, 9)

	uints := make([]uint, 10)
	for i := 0; i < 10; i++ {
		uints[i] = uint(rand.Uint32())
		encodeUInt(&buf, uints[i], temp)
	}
	reader := bytes.NewReader(buf.Bytes())
	for i := 0; i < 10; i++ {
		j := decodeUInt(reader, temp)
		if uints[i] != j {
			t.Fatal("Incorrect int recovered")
		}
	}
}

func TestEncodeFloat(t *testing.T) {
	var buf bytes.Buffer
	temp := make([]byte, 9)
	floats := make([]float64, 10)
	for i := 0; i < 10; i++ {
		floats[i] = rand.Float64()
		encodeFloat(&buf, floats[i], temp)
	}
	reader := bytes.NewReader(buf.Bytes())
	for i := 0; i < 10; i++ {
		j := decodeFloat(reader, temp)
		if floats[i] != j {
			t.Fatal("Incorrect float recovered")
		}
	}
}

func TestEncodeStringSlice(t *testing.T) {
	var buf bytes.Buffer
	temp := make([]byte, 9)
	encodeStringSlice(&buf, testWords, temp)
	reader := bytes.NewReader(buf.Bytes())
	unserialized := decodeStringSlice(reader, temp)
	for i, w := range unserialized {
		if testWords[i] != w {
			t.Fatal("Incorrect float recovered")
		}
	}
}

func TestEncodeTrie(t *testing.T) {
	trie := NewTrie()
	for i, w := range testWords {
		var wf weights
		wf[0] = float64(i)
		trie.add(w, i, wf)
	}
	trie.Serialize("test")
	//defer os.Remove(path.Join("indexes", "test.index"))
	unserialized := UnserializeTrie("test")
	for _, w := range testWords {
		resp := unserialized.get(w)
		if len(resp) != 1 {
			t.Fatal("Incorrect result size for inserted word")
		}
	}
	for _, w := range fakeWords {
		if len(trie.get(w)) != 0 {
			t.Fatal("Incorrect word found")
		}
	}
}
