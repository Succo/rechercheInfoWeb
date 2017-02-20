package main

import "testing"

var testWords = []string{
	"abjure",
	"abrogate",
	"abstemious",
	"acumen",
	"antebellum",
	"auspicious",
	"belie",
	"bellicose",
	"bowdlerize",
	"chicanery",
	"chromosome",
	"churlish",
	"circumlocution",
	"circumnavigate",
	"deciduous",
	"deleterious",
	"diffident",
	"enervate",
	"enfranchise",
	"epiphany",
	"equinox",
	"euro",
	"evanescent",
	"expurgate",
	"facetious",
	"fatuous",
	"feckless",
	"fiduciary",
	"filibuster",
	"gamete",
	"gauche",
	"gerrymander",
}

var fakeWords = []string{
	"hemoglobin",
	"homogeneous",
	"hubris",
	"hypotenuse",
	"impeach",
	"incognito",
	"incontrovertible",
	"inculcate",
	"infrastructure",
	"interpolate",
	"irony",
	"jejune",
	"kinetic",
	"kowtow",
	"laissez faire",
	"lexicon",
	"loquacious",
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
		trie.add(w, i, wf)
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
