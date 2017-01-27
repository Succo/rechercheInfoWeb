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
	testRefs := make([]*Ref, len(testWords))
	trie := NewTrie()
	for i, w := range testWords {
		ref := &Ref{}
		testRefs[i] = ref
		trie.add(w, ref)
	}
	for i, w := range testWords {
		for _, ref := range trie.get(w) {
			if ref != testRefs[i] {
				t.Fail()
			}
		}
	}
	for _, w := range fakeWords {
		if len(trie.get(w)) != 0 {
			t.Fail()
		}
	}

}
