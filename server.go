package main

import (
	"html/template"
	"math"
	"net/http"
)

type answer struct {
	Results []string
}

type stat struct {
	Name       string
	Documents  int
	Vocabulary int
	Tokens     int
	B          float64
	K          float64
}

func getStat(s *Search, name string) stat {
	corpusSize := s.CorpusSize()
	tokenSize := s.TokenSize(corpusSize)
	halfTokenSize := s.TokenSize(corpusSize / 2)
	vocabSize := s.IndexSize(corpusSize)
	halfVocabSize := s.IndexSize(corpusSize / 2)

	// Heaps law calculation
	b := (math.Log(float64(tokenSize)) - math.Log(float64(halfTokenSize))) / (math.Log(float64(vocabSize)) - math.Log(float64(halfVocabSize)))
	k := float64(tokenSize) / (math.Pow(float64(vocabSize), b))

	return stat{
		Name:       name,
		Documents:  corpusSize,
		Tokens:     tokenSize,
		Vocabulary: vocabSize,
		B:          b,
		K:          k,
	}
}

func serve(cacm, cs276 *Search) {
	index, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err.Error())
	}

	statT, err := template.ParseFiles("templates/stat.html")
	if err != nil {
		panic(err.Error())
	}
	stats := make([]stat, 0, 2)
	for i, search := range []*Search{cacm, cs276} {
		// small hack, name should part of the search struct
		var name string
		if i == 0 {
			name = "cacm"
		} else {
			name = "cs276"
		}

		stats = append(stats, getStat(search, name))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		corpus := r.FormValue("corpus")
		input := r.FormValue("search")
		if len(corpus) == 0 || len(input) == 0 {
			index.Execute(w, nil)
			return
		}

		var search *Search
		if corpus == "cacm" {
			search = cacm
		} else if corpus == "cs276" {
			search = cs276
		} else {
			index.Execute(w, nil)
			return
		}
		results := search.Search(input)

		index.Execute(w, answer{Results: results})
	})

	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		statT.Execute(w, stats)
	})

	http.HandleFunc("/cacm.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cacm.svg")
	})

	http.HandleFunc("/cs276.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cs276.svg")
	})

	http.ListenAndServe(":8080", nil)
}
