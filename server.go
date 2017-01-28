package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type answer struct {
	Query string
	// A small "hack" to keep the button checked
	CS276   bool
	Results []Result
	Time    string
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
	stats := []*Stat{
		&cacm.Stat,
		&cs276.Stat,
	}

	cacmT, err := template.ParseFiles("templates/cacm.html")
	if err != nil {
		panic(err.Error())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		corpus := r.FormValue("corpus")
		input := r.FormValue("search")
		if len(corpus) == 0 || len(input) == 0 {
			index.Execute(w, nil)
			return
		}

		var search *Search
		a := answer{Query: input}
		if corpus == "cacm" {
			search = cacm
		} else if corpus == "cs276" {
			search = cs276
			a.CS276 = true
		} else {
			index.Execute(w, nil)
			return
		}
		now := time.Now()
		a.Results = search.Search(input)
		a.Time = time.Since(now).String()

		index.Execute(w, a)
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

	log.Println("riw starting to serve traffic")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
