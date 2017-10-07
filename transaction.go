package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrrpcclient"
)

type ticketPurchaseTransaction struct {
}

type coinbaseTransaction struct {
}

type regularTransaction struct {
}

type revocationTransaction struct {
}

type voteTransaction struct {
}

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
		GetRawTransactionVerbose, err := client.GetRawTransactionVerbose(transactionHash)

		if err != nil {
			log.Fatal(err)
		}

		transactionType := GetRawTransactionVerbose.Vout[0].ScriptPubKey.Type

		if transactionType == "stakesubmission" {
			// Ticket Purchase
		} else if transactionType == "scripthash" {
			// Coinbase
		} else if transactionType == "pubkeyhash" {
			// Regular Transaction
		} else if transactionType == "stakerevoke" {
			// Revocation
		} else {
			// Vote
		}
		t.Execute(w, GetRawTransactionVerbose)
	}
}
