package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrrpcclient"
)

func handleTransaction(w http.ResponseWriter, r *http.Request, client *dcrrpcclient.Client) {
	t, err := template.ParseFiles("templates/transaction.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, "")
}
