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
	Hex           string
	TxID          string
	Version       int32
	LockTime      uint32
	Expiry        uint32
	Inputs        []txnInput
	Outputs       []txnOutput
	BlockHash     string
	BlockHeight   int64
	Confirmations int64
	Time          time.Time
	Type          string
}

type txnInput struct {
	Coinbase    string
	Stakebase   string
	TxID        string
	Vout        uint32
	Tree        int8
	Sequence    uint32
	AmountIn    float64
	BlockHeight uint32
	BlockIndex  uint32
	ScriptSig   *dcrjson.ScriptSig
}

type txnOutput struct {
	Value        float64
	N            uint32
	Version      uint16
	ScriptPubKey dcrjson.ScriptPubKeyResult
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
		rawTransactionVerbose, err := client.GetRawTransactionVerbose(transactionHash)

		if err != nil {
			log.Fatal(err)
		}

		displayTransaction := new(displayTransaction)

		transactionType := rawTransactionVerbose.Vout[0].ScriptPubKey.Type

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
		} else if rawTransactionVerbose.Vout[2].ScriptPubKey.Type == "stakegen" {
			// Vote
			displayTransaction.Type = "Vote"
		}

		t.Execute(w, rawTransactionVerbose)
	}
}
