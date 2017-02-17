package main

import "testing"

var testWords = [][]byte{
	[]byte("abjure"),
	[]byte("abrogate"),
	[]byte("abstemious"),
	[]byte("acumen"),
	[]byte("antebellum"),
	[]byte("auspicious"),
	[]byte("belie"),
	[]byte("bellicose"),
	[]byte("bowdlerize"),
	[]byte("chicanery"),
	[]byte("chromosome"),
	[]byte("churlish"),
	[]byte("circumlocution"),
	[]byte("circumnavigate"),
	[]byte("deciduous"),
	[]byte("deleterious"),
	[]byte("diffident"),
	[]byte("enervate"),
	[]byte("enfranchise"),
	[]byte("epiphany"),
	[]byte("equinox"),
	[]byte("euro"),
	[]byte("evanescent"),
	[]byte("expurgate"),
	[]byte("facetious"),
	[]byte("fatuous"),
	[]byte("feckless"),
	[]byte("fiduciary"),
	[]byte("filibuster"),
	[]byte("gamete"),
	[]byte("gauche"),
	[]byte("gerrymander"),
}

var fakeWords = [][]byte{
	[]byte("hemoglobin"),
	[]byte("homogeneous"),
	[]byte("hubris"),
	[]byte("hypotenuse"),
	[]byte("impeach"),
	[]byte("incognito"),
	[]byte("incontrovertible"),
	[]byte("inculcate"),
	[]byte("infrastructure"),
	[]byte("interpolate"),
	[]byte("irony"),
	[]byte("jejune"),
	[]byte("kinetic"),
	[]byte("kowtow"),
	[]byte("laissez faire"),
	[]byte("lexicon"),
	[]byte("loquacious"),
}

func TestTrie(t *testing.T) {
	testDeltas := make([]int, len(testWords))
	testTfIds := make([]weights, len(testWords))
	trie := NewTrie()
	for i, w := range testWords {
		testDeltas[i] = int(i)
		testTfIds[i][0] = float64(i)
		var wf weights
		wf[0] = float64(i)
		trie.add(w, uint(i), wf)
	}
	for i, w := range testWords {
		resp := trie.get(w)
		if len(resp) != 1 {
			t.Fatal("Incorrect result size for inserted word")
		} else if resp[0].Id != testDeltas[i] ||
			resp[0].Weights != testTfIds[i] {
			t.Fatal("Incorrect result for inserted word")
		}
	}
	for _, w := range fakeWords {
		if len(trie.get(w)) != 0 {
			t.Fatal("Incorrect word found")
		}
	}
}
