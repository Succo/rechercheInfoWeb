package main

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestEncodeInt(t *testing.T) {
	var buf bytes.Buffer
	ints := make([]int, 10)
	for i := 0; i < 10; i++ {
		ints[i] = rand.Int()
		encodeInt(&buf, ints[i])
	}
	reader := bytes.NewReader(buf.Bytes())
	for i := 0; i < 10; i++ {
		j := decodeInt(reader)
		if ints[i] != j {
			t.Fatal("Incorrect int recovered")
		}
	}
}

func TestEncodeFloat(t *testing.T) {
	var buf bytes.Buffer
	floats := make([]float64, 10)
	for i := 0; i < 10; i++ {
		floats[i] = rand.Float64()
		encodeFloat(&buf, floats[i])
	}
	reader := bytes.NewReader(buf.Bytes())
	for i := 0; i < 10; i++ {
		j := decodeFloat(reader)
		if floats[i] != j {
			t.Fatal("Incorrect float recovered")
		}
	}
}

func TestEncodeStringSlice(t *testing.T) {
	var buf bytes.Buffer
	encodeStringSlice(&buf, testWords)
	reader := bytes.NewReader(buf.Bytes())
	unserialized := decodeStringSlice(reader)
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
