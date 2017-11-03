package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrutil"
)

type displayAddress struct {
	SearchRawTransactionsResult []*dcrjson.SearchRawTransactionsResult
	TransactionCount            int
	TotalSent                   float64
	TotalReceived               float64
	FinalBalance                float64
}

func handleAddress(w http.ResponseWriter, r *http.Request, client *rpcclient.Client) {

	t, err := template.ParseFiles("templates/address.html", "templates/partial_head.html")
	if err != nil {
		log.Fatal(err)
	}

	inputAddressStr := r.URL.Path[9:]
	if len(inputAddressStr) != 35 {
		w.Write([]byte("Error: Address must be 35 characters"))
	} else {
		inputAddress, err := dcrutil.DecodeAddress(inputAddressStr)
		if err != nil {
			log.Fatal(err)
		}
		displayAddress := new(displayAddress)
		displayAddress.SearchRawTransactionsResult, err = client.SearchRawTransactionsVerbose(inputAddress, 0, 100, true, false, nil)
		if err != nil {
			log.Fatal(err)
		}

		t.Execute(w, displayAddress)
	}
}
