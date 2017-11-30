package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/decred/dcrd/dcrjson"

	"github.com/decred/dcrd/rpcclient"
)

type indexBlocks struct {
	VerboseBlock *dcrjson.GetBlockVerboseResult
	Age          int
	TxnCount     int
	VoteCount    int
	TicketCount  int
}

func handleIndex(w http.ResponseWriter, r *http.Request, client *rpcclient.Client) {

	t, err := template.ParseFiles("templates/index.html", "templates/partial_head.html")
	if err != nil {
		log.Fatal(err)
	}

	currentBlockHeight, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}

	var topTenBlocks [10]indexBlocks

	for i := 0; i < 10; i++ {

		blockHash, err := client.GetBlockHash(currentBlockHeight - int64(i))
		if err != nil {
			log.Fatal(err)
		}

		topTenBlocks[i].VerboseBlock, err = client.GetBlockVerbose(blockHash, true)
		if err != nil {
			log.Fatal(err)
		}

	}

	// Get the current block height
	// Create structs with the last 10 blocks

	t.Execute(w, topTenBlocks)

}
