package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/decred/dcrd/rpcclient"
)

type indexBlocks struct {
	Height      int64
	Size        int32
	Age         float64
	TxnCount    int
	VoteCount   int
	TicketCount int
	RevokeCount int
	Reward      float64
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

		verboseBlock, err := client.GetBlockVerbose(blockHash, true)
		if err != nil {
			log.Fatal(err)
		}

		topTenBlocks[i].Height = verboseBlock.Height
		topTenBlocks[i].Size = verboseBlock.Size

		topTenBlocks[i].Age = math.Floor(time.Now().Sub(time.Unix(verboseBlock.Time, 0)).Minutes())

		topTenBlocks[i].TxnCount = len(verboseBlock.RawTx)

		topTenBlocks[i].VoteCount = 0
		topTenBlocks[i].TicketCount = 0
		topTenBlocks[i].RevokeCount = 0

		for j := 0; j < len(verboseBlock.RawSTx); j++ {
			if verboseBlock.RawSTx[j].Vout[0].ScriptPubKey.Type == "stakesubmission" {
				topTenBlocks[i].TicketCount++
			} else if verboseBlock.RawSTx[j].Vout[0].ScriptPubKey.Type == "stakerevoke" {
				topTenBlocks[i].RevokeCount++
			} else if verboseBlock.RawSTx[j].Vout[2].ScriptPubKey.Type == "stakegen" {
				topTenBlocks[i].VoteCount++
			}
		}

		for h := 0; h < len(verboseBlock.RawTx); h++ {
			if verboseBlock.RawTx[h].Vin[0].IsCoinBase() {
				topTenBlocks[i].Reward = verboseBlock.RawTx[h].Vin[0].AmountIn
			}
		}

	}

	// Get the current block height
	// Create structs with the last 10 blocks

	t.Execute(w, topTenBlocks)

}
