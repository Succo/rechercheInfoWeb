package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
)

const (
	maxSize = 20
)

type answer struct {
	Query string
	// A small "hack" to keep the same buttons checked
	CS276     bool
	Vectorial bool
	Norm      bool
	Results   []Result
	Time      string
	// Links to other results in the query set
	Prev string
	Next string
	Size int
}

func printDuration(dur time.Duration) string {
	// Round it to a ms first
	return ((dur / time.Millisecond) * time.Millisecond).String()
}

func serve(cacm, cs276 *Search) {
	prettyfier := template.FuncMap{
		"duration": printDuration,
		"size":     humanize.Bytes,
	}
	index, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err.Error())
	}

	statT, err := template.ParseFiles("templates/stat.html")
	if err != nil {
		panic(err.Error())
	}
	stats := []*Stat{
		&cacm.Stat,
		&cs276.Stat,
	}

	perfT, err := template.New("perf.html").Funcs(prettyfier).ParseFiles("templates/perf.html")
	if err != nil {
		panic(err.Error())
	}
	perfs := []*Perf{
		&cacm.Perf,
		&cs276.Perf,
	}

	cacmT, err := template.ParseFiles("templates/cacm.html")
	if err != nil {
		panic(err.Error())
	}

	// Histogram used for monitoring of search time
	cacmH := expvar.NewHistogram("cacm", 50)
	cs276H := expvar.NewHistogram("cs276", 50)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		corpus := r.FormValue("corpus")
		input := r.FormValue("search")
		if len(corpus) == 0 || len(input) == 0 {
			index.Execute(w, nil)
			return
		}

		searchType := r.FormValue("type")

		var offset int
		offset, err := strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			offset = 0
		}

		var search *Search
		var hist metrics.Histogram
		a := answer{Query: input}
		if corpus == "cacm" {
			search = cacm
			hist = cacmH
		} else if corpus == "cs276" {
			search = cs276
			hist = cs276H
			a.CS276 = true
		} else {
			index.Execute(w, nil)
			return
		}
		now := time.Now()
		if searchType == "boolean" {
			a.Results = search.BooleanSearch(input)
		} else if searchType == "vectorial" {
			weightFun := r.FormValue("weight")
			if weightFun == "norm" {
				a.Results = search.VectorSearch(input, 1)
				a.Norm = true
			} else {
				a.Results = search.VectorSearch(input, 0)
			}
			a.Vectorial = true
		} else {
			index.Execute(w, nil)
			return
		}
		hist.Observe(float64(time.Since(now)))
		a.Time = time.Since(now).String()
		a.Size = len(a.Results)
		if offset > 0 && len(a.Results) > offset {
			a.Results = a.Results[offset:]
			if offset > 0 {
				a.Prev = fmt.Sprintf("/?search=%s&offset=%d&corpus=%s&type=%s",
					input, max(offset-maxSize, 0), corpus, searchType)
			}
		} else {
			offset = 0
		}
		if len(a.Results) > maxSize {
			a.Results = a.Results[:maxSize]
			a.Next = fmt.Sprintf("/?search=%s&offset=%d&corpus=%s&type=%s",
				input, offset+maxSize, corpus, searchType)
		}

		index.Execute(w, a)
	})

	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		statT.Execute(w, stats)
	})

	http.HandleFunc("/perf", func(w http.ResponseWriter, r *http.Request) {
		err := perfT.Execute(w, perfs)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/torch.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "torch.svg")
	})

	http.HandleFunc("/cacm.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphs/cacm.svg")
	})

	http.HandleFunc("/cs276.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphs/cs276.svg")
	})

	http.HandleFunc("/cacm/", func(w http.ResponseWriter, r *http.Request) {
		// len("/cacm/") = 5
		id, err := strconv.Atoi(r.URL.Path[6:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		doc, err := getCACMDoc(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		cacmT.Execute(w, doc)
	})

	http.HandleFunc("/percentile", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/percentile.html")
	})

	log.Println("riw starting to serve traffic")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
