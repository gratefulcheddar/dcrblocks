package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrrpcclient"
)

func handleIndex(w http.ResponseWriter, r *http.Request, client *dcrrpcclient.Client) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, "")
}
