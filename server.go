package main

import (
	"html/template"
	"net/http"
)

func serve() {
	index, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err.Error())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index.Execute(w, nil)
	})

	http.ListenAndServe(":8080", nil)
}
