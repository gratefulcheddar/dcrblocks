package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrrpcclient"
)

func handleTransaction(w http.ResponseWriter, r *http.Request, client *dcrrpcclient.Client) {
	t, err := template.ParseFiles("templates/transaction.html")
	if err != nil {
		log.Fatal(err)
	}

	inputTxnStr := r.URL.Path[13:]
	if len(inputTxnStr) != 64 {
		w.Write([]byte("Error: Transaction ID must be 64 characters"))
	} else {

		transactionHash, err := chainhash.NewHashFromStr(inputTxnStr)

		if err != nil {
			log.Fatal(err)
		}
		rawTransactionVerbose, err := client.GetRawTransactionVerbose(transactionHash)

		if err != nil {
			log.Fatal(err)
		}

		// transactionType := rawTransactionVerbose.Vout[0].ScriptPubKey.Type

		/*var typeString string

		if transactionType == "stakesubmission" {
			// Ticket Purchase
			typeString = "Ticket Purchase"
		} else if transactionType == "scripthash" {
			// Coinbase
			typeString = "Coinbase"
		} else if transactionType == "pubkeyhash" {
			// Regular Transaction
			typeString = "Transaction"
		} else if transactionType == "stakerevoke" {
			// Revocation
			typeString = "Revocation"
		} else {
			// Vote
			typeString = "Vote"
		}*/

		t.Execute(w, rawTransactionVerbose)
	}
}
