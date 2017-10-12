package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrrpcclient"
)

type displayTransaction struct {
	RawTransactionVerbose *dcrjson.TxRawResult
	Time                  time.Time
	Type                  string
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

		displayTransaction := new(displayTransaction)

		displayTransaction.RawTransactionVerbose, err = client.GetRawTransactionVerbose(transactionHash)

		if err != nil {
			log.Fatal(err)
		}

		transactionType := displayTransaction.RawTransactionVerbose.Vout[0].ScriptPubKey.Type

		if transactionType == "stakesubmission" {
			// Ticket Purchase
			displayTransaction.Type = "Ticket Purchase"
		} else if transactionType == "scripthash" {
			// Coinbase
			displayTransaction.Type = "Coinbase"
		} else if transactionType == "pubkeyhash" {
			// Regular Transaction
			displayTransaction.Type = "Transaction"
		} else if transactionType == "stakerevoke" {
			// Revocation
			displayTransaction.Type = "Revocation"
		} else if displayTransaction.RawTransactionVerbose.Vout[2].ScriptPubKey.Type == "stakegen" {
			// Vote
			displayTransaction.Type = "Vote"
		}

		displayTransaction.Time = time.Unix(displayTransaction.RawTransactionVerbose.Time, 0)

		t.Execute(w, displayTransaction)
	}
}
