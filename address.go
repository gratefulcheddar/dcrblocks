package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrutil"
)

type displayAddress struct {
	AddressString    string
	TransactionCount int
	TotalSent        float64
	TotalReceived    float64
	FinalBalance     float64
	Sent             []*basicDisplayTransaction
	Received         []*basicDisplayTransaction
}
type basicDisplayTransaction struct {
	Amount      *float64
	Txid        string
	BlockHeight uint64
}

func handleAddress(w http.ResponseWriter, r *http.Request, client *rpcclient.Client) {

	t, err := template.ParseFiles("templates/address.html", "templates/partial_head.html", "templates/vin_table.html")
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
		displayAddress.AddressString = inputAddressStr
		searchRawTransactionsResult, err := client.SearchRawTransactionsVerbose(inputAddress, 0, 100, true, false, nil)
		if err != nil {
			log.Fatal(err)
		}
		for _, transaction := range searchRawTransactionsResult {

			basicTransaction := new(basicDisplayTransaction)
			basicTransaction.Txid = transaction.Txid

			currentBlockHeight, err := client.GetBlockCount()
			if err != nil {
				log.Fatal(err)
			}

			basicTransaction.BlockHeight = uint64(currentBlockHeight) - transaction.Confirmations

			for vinIndex, vin := range transaction.Vin {
				if vin.PrevOut != nil {
					for _, address := range vin.PrevOut.Addresses {
						if address == inputAddressStr {
							basicTransaction.Amount = &transaction.Vin[vinIndex].PrevOut.Value
							displayAddress.Sent = append(displayAddress.Sent, basicTransaction)
						}
					}
				}
			}

			for voutIndex, vout := range transaction.Vout {
				for _, address := range vout.ScriptPubKey.Addresses {
					if address == inputAddressStr {
						basicTransaction.Amount = &transaction.Vout[voutIndex].Value
						displayAddress.Received = append(displayAddress.Received, basicTransaction)
					}
				}
			}

			displayAddress.TransactionCount = len(displayAddress.Sent) + len(displayAddress.Received)

			sentSum := 0.0

			for _, sentTxn := range displayAddress.Sent {
				sentSum += *sentTxn.Amount
			}

			displayAddress.TotalSent = sentSum

			receivedSum := 0.0

			for _, receivedTxn := range displayAddress.Received {
				receivedSum += *receivedTxn.Amount
			}

			displayAddress.TotalReceived = receivedSum

			displayAddress.FinalBalance = displayAddress.TotalReceived - displayAddress.TotalSent
		}

		t.Execute(w, displayAddress)
	}
}
