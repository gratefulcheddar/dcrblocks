// Copyright (c) 2014-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrrpcclient"
	"github.com/decred/dcrutil"
)

type DisplayBlock struct {
	// Versions:
	BlockVersion int32
	VoteVersion  uint32

	// Summary:
	BlockHeight      int64
	TransactionCount int
	Confirmations    int64
	BlockSize        int32
	Timestamp        time.Time
	NextBlock        int64
	PreviousBlock    int64

	// Hashes:
	Hash       string
	MerkleRoot string
	StakeRoot  string

	// Tickets:
	VotesCast        uint16
	TicketPrice      float64
	TicketsPurchased uint8
	Revocations      uint8
	TicketPoolSize   uint32

	Transactions    []*dcrjson.TxRawResult
	Votes           []*dcrjson.TxRawResult
	TicketPurchases []*dcrjson.TxRawResult
}

func handleBlock(w http.ResponseWriter, r *http.Request, client *dcrrpcclient.Client) {

	// Reads the block number as a string from the URL Path
	inputBlockStr := r.URL.Path[7:]

	// Convert the block number string to an int64
	inputBlockInt64, err := strconv.ParseInt(inputBlockStr, 10, 64)
	if err != nil {
		w.Write([]byte("Error: Block height contains invalid character"))
	} else {

		// Get the current block height
		currentBlockHeight, err := client.GetBlockCount()
		if err != nil {
			log.Fatal(err)
		}

		// Check that input block is valid
		if inputBlockInt64 > currentBlockHeight {
			w.Write([]byte("Error: Input block exceeds current chain height."))
		} else {

			// Get the block hash of the input block
			blockHash, err := client.GetBlockHash(inputBlockInt64)
			if err != nil {
				log.Print(err)
			}

			// Parse the block.html template
			t, err := template.ParseFiles("templates/block.html")
			if err != nil {
				log.Fatal(err)
			}

			block, err := client.GetBlockVerbose(blockHash, true)
			if err != nil {
				log.Fatal(err)
			}

			displayBlock := new(DisplayBlock)
			displayBlock.BlockHeight = block.Height
			displayBlock.BlockSize = block.Size
			displayBlock.BlockVersion = block.Version
			displayBlock.Confirmations = currentBlockHeight - int64(block.Height)
			displayBlock.Hash = blockHash.String()
			displayBlock.MerkleRoot = block.MerkleRoot
			displayBlock.NextBlock = block.Height + 1
			displayBlock.PreviousBlock = block.Height - 1
			displayBlock.Revocations = block.Revocations
			displayBlock.StakeRoot = block.StakeRoot
			displayBlock.TicketPoolSize = block.PoolSize
			displayBlock.TicketPrice = block.SBits
			displayBlock.TicketsPurchased = block.FreshStake
			displayBlock.Timestamp = time.Unix(block.Time, 0)
			displayBlock.TransactionCount = len(block.Tx)
			displayBlock.VotesCast = block.Voters
			displayBlock.VoteVersion = block.StakeVersion

			for i := 0; i < len(block.RawTx); i++ {
				displayBlock.Transactions = append(displayBlock.Transactions, &block.RawTx[i])
			}

			for i := 0; i < len(block.RawSTx); i++ {
				if block.RawSTx[i].Vout[0].ScriptPubKey.Type == "stakesubmission" {
					displayBlock.TicketPurchases = append(displayBlock.TicketPurchases, &block.RawSTx[i])
				} else {
					displayBlock.Votes = append(displayBlock.Votes, &block.RawSTx[i])
				}

			}

			// Load the template with the block data
			t.Execute(w, displayBlock)
		}
	}
}

// TODO:
// parseBlock parses a MsgBlock into a DisplayBlock for use
// with the block.html templates

// makeHandler creates a new dcrrpcclient, passes it to the
// input function for executing, shuts down the client,
// and returns an http.HandlerFunc to use with http.HandleFunc().
func makeHandler(fn func(http.ResponseWriter, *http.Request, *dcrrpcclient.Client)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Connect to local dcrd RPC server using websockets.
		dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
		certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
		if err != nil {
			log.Fatal(err)
		}
		connCfg := &dcrrpcclient.ConnConfig{
			Host:         "localhost:9109",
			Endpoint:     "ws",
			User:         "cheesepool",
			Pass:         "dcrblocks",
			Certificates: certs,
		}
		client, err := dcrrpcclient.New(connCfg, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Pass client to input function, fn.
		fn(w, r, client)

		// Shutdown the client connection.
		client.Shutdown()
	}
}

func main() {
	// Serve static files (css, images, etc)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	server := http.Server{
		Addr: "localhost:8080",
	}

	http.HandleFunc("/block/", makeHandler(handleBlock))
	log.Printf("Server started. Address = %v", server.Addr)
	server.ListenAndServe()

}
