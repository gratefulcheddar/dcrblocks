package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/rpcclient"
)

type displayTransaction struct {
	RawTransactionVerbose *dcrjson.TxRawResult
	Time                  time.Time
	Type                  string
	Votes                 *parsedVote
}

func handleTransaction(w http.ResponseWriter, r *http.Request, client *rpcclient.Client) {
	t, err := template.ParseFiles("templates/transaction.html", "templates/partial_head.html")
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
			displayTransaction.Type = "Regular Transaction"
		} else if transactionType == "stakerevoke" {
			// Revocation
			displayTransaction.Type = "Revocation"
		} else if displayTransaction.RawTransactionVerbose.Vout[2].ScriptPubKey.Type == "stakegen" {
			// Vote
			displayTransaction.Type = "Vote"
			displayTransaction.Votes = parseVoteScript(displayTransaction.RawTransactionVerbose.Vout[1].ScriptPubKey.Hex)
		}

		displayTransaction.Time = time.Unix(displayTransaction.RawTransactionVerbose.Time, 0)

		t.Execute(w, displayTransaction)
	}
}
