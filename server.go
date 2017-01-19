package main

import (
	"html/template"
	"net/http"
)

type answer struct {
	Results []string
}

func serve(cacm, cs276 *Search) {
	index, err := template.ParseFiles("templates/index.html")
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

	http.ListenAndServe(":8080", nil)
}
