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
	testRefs := make([]Ref, len(testWords))
	trie := NewTrie()
	for i, w := range testWords {
		ref := Ref{}
		testRefs[i] = ref
		trie.add(w, ref)
	}
	for i, w := range testWords {
		resp := trie.get(w)
		if len(resp) != 1 {
			t.Fatal("Incorrect result size for inserted word")
		} else if resp[0] != testRefs[i] {
			t.Fatal("Incorrect result for inserted word")
		}
	}
	for _, w := range fakeWords {
		if len(trie.get(w)) != 0 {
			t.Fatal("Incorrect word found")
		}
	}
}
