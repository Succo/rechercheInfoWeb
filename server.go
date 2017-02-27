package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
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
	Weight    string
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

func serve(cacm, cs276 *Search, precall *PreCallCalculator) {
	prettyfier := template.FuncMap{
		"duration": printDuration,
		"size":     humanize.Bytes,
	}

	pattern := path.Join("templates", "*.html")
	templates := template.Must(template.New("base").Funcs(prettyfier).ParseGlob(pattern))

	stats := []*Stat{
		&cacm.Stat,
		&cs276.Stat,
	}
	perfs := []*Perf{
		&cacm.Perf,
		&cs276.Perf,
	}

	// Histogram used for monitoring of search time
	cacmH := expvar.NewHistogram("cacm", 50)
	cs276H := expvar.NewHistogram("cs276", 50)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		corpus := r.FormValue("corpus")
		input := r.FormValue("search")
		if len(corpus) == 0 || len(input) == 0 {
			templates.ExecuteTemplate(w, "index", answer{})
			return
		}

		searchType := r.FormValue("type")
		weightFun := r.FormValue("weight")

		var offset int
		offset, _ = strconv.Atoi(r.FormValue("offset"))

		var search *Search
		var hist metrics.Histogram
		a := answer{Query: input, Weight: weightFun}
		if corpus == "cacm" {
			search = cacm
			hist = cacmH
		} else if corpus == "cs276" {
			search = cs276
			hist = cs276H
			a.CS276 = true
		} else {
			templates.ExecuteTemplate(w, "index", a)
			return
		}
		now := time.Now()
		if searchType == "boolean" {
			a.Results = search.BooleanSearch(input)
		} else if searchType == "vectorial" {
			if weightFun == "norm" {
				a.Results = search.VectorSearch(input, norm)
			} else if weightFun == "half" {
				a.Results = search.VectorSearch(input, half)
			} else {
				a.Results = search.VectorSearch(input, raw)
			}
			a.Vectorial = true
		} else {
			templates.ExecuteTemplate(w, "index", a)
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

		templates.ExecuteTemplate(w, "index", a)
	})

	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "stat", stats)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/perf", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "perf", perfs)
		if err != nil {
			log.Fatal(err.Error())
		}
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
		err = templates.ExecuteTemplate(w, "cacm", doc)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/qrels", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "qrels", precall)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/precall", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "precall", precall)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/archi", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "archi", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	// Static content
	fs := http.FileServer(http.Dir("graphs"))
	http.Handle("/graphs/", http.StripPrefix("/graphs/", fs))

	http.HandleFunc("/percentile", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "percentile", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
	})

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	http.HandleFunc("/archi_indexing.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "archi_indexing.svg")
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
