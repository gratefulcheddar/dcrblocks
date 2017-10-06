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
	RevocationCount  uint8
	TicketPoolSize   uint32
	VoteReward       float64

	Transactions    []RegularTransaction
	Votes           []VoteTransaction
	TicketPurchases []TicketPurchaseTransaction
	Revocations     []RevocationTransaction
}

type RegularTransaction struct {
	Amount float64
	TxID   string
}

type VoteTransaction struct {
	Votes   map[string]string
	TxID    string
	Version string
}

type TicketPurchaseTransaction struct {
	TxID     string
	Maturity string
}

type RevocationTransaction struct {
	TxID string
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

			blockSubsidy, err := client.GetBlockSubsidy(block.Height, block.Voters)
			if err != nil {
				log.Fatal(err)
			}

			// Main purpose of creating a new struct is to allow the filtering of
			// transactions by type. This can be moved to a separate function...
			displayBlock := new(DisplayBlock)
			displayBlock.BlockHeight = block.Height
			displayBlock.BlockSize = block.Size
			displayBlock.BlockVersion = block.Version
			displayBlock.Confirmations = block.Confirmations
			displayBlock.Hash = block.Hash
			displayBlock.MerkleRoot = block.MerkleRoot
			if displayBlock.BlockHeight != currentBlockHeight {
				displayBlock.NextBlock = block.Height + 1
			}
			displayBlock.PreviousBlock = block.Height - 1
			displayBlock.RevocationCount = block.Revocations
			displayBlock.StakeRoot = block.StakeRoot
			displayBlock.TicketPoolSize = block.PoolSize
			displayBlock.TicketPrice = block.SBits
			displayBlock.TicketsPurchased = block.FreshStake
			displayBlock.Timestamp = time.Unix(block.Time, 0)
			displayBlock.TransactionCount = len(block.RawTx)
			displayBlock.VotesCast = block.Voters
			displayBlock.VoteReward = (float64(blockSubsidy.PoS) / 100000000) / float64(block.Voters)
			displayBlock.VoteVersion = block.StakeVersion

			// Loop through the Raw Transactions, creating a slice of
			// RegularTransactions with the minimum information
			// required to display to user.
			for i := 0; i < len(block.RawTx); i++ {
				newTransaction := new(RegularTransaction)
				newTransaction.TxID = block.RawTx[i].Txid
				for _, value := range block.RawTx[i].Vin {
					newTransaction.Amount += value.AmountIn
				}
				if block.Height == 0 {
					newTransaction.Amount = 0
				}
				displayBlock.Transactions = append(displayBlock.Transactions, *newTransaction)
			}

			// Loop through the Raw Stake Transactions, creating two
			// slices, one of TicketPurchaseTransactions and one of
			// VoteTransactions, with the minimum information required
			// to display to user.
			for i := 0; i < len(block.RawSTx); i++ {
				if block.RawSTx[i].Vout[0].ScriptPubKey.Type == "stakesubmission" {
					ticketPurchase := new(TicketPurchaseTransaction)
					ticketPurchase.TxID = block.RawSTx[i].Txid

					if block.RawSTx[i].Confirmations > 256 {
						ticketPurchase.Maturity = "Mature"
					} else {
						ticketPurchase.Maturity = "Immature"
					}

					displayBlock.TicketPurchases = append(displayBlock.TicketPurchases, *ticketPurchase)

				} else if block.RawSTx[i].Vout[0].ScriptPubKey.Type == "stakerevoke" {
					revocation := new(RevocationTransaction)
					revocation.TxID = block.RawSTx[i].Txid
					displayBlock.Revocations = append(displayBlock.Revocations, *revocation)
				} else {
					vote := new(VoteTransaction)
					vote.TxID = block.RawSTx[i].Txid
					vote.Votes = make(map[string]string)
					// Parse Vote - TODO: Make this automatic
					if len(block.RawSTx[i].Vout[1].ScriptPubKey.Hex) > 8 {
						switch block.RawSTx[i].Vout[1].ScriptPubKey.Hex[8:10] {
						case "04":
							vote.Version = "4"
							switch block.RawSTx[i].Vout[1].ScriptPubKey.Hex[4:6] {
							case "00":
								fallthrough
							case "01":
								vote.Votes["lnsupport"] = "abstain"
								vote.Votes["sdiffalgo"] = "abstain"
							case "02":
								fallthrough
							case "03":
								vote.Votes["lnsupport"] = "abstain"
								vote.Votes["sdiffalgo"] = "no"
							case "04":
								fallthrough
							case "05":
								vote.Votes["lnsupport"] = "abstain"
								vote.Votes["sdiffalgo"] = "yes"
							case "08":
								fallthrough
							case "09":
								vote.Votes["lnsupport"] = "no"
								vote.Votes["sdiffalgo"] = "abstain"
							case "0a":
								fallthrough
							case "0b":
								vote.Votes["lnsupport"] = "no"
								vote.Votes["sdiffalgo"] = "no"
							case "0c":
								fallthrough
							case "0d":
								vote.Votes["lnsupport"] = "no"
								vote.Votes["sdiffalgo"] = "yes"
							case "10":
								fallthrough
							case "11":
								vote.Votes["lnsupport"] = "yes"
								vote.Votes["sdiffalgo"] = "abstain"
							case "12":
								fallthrough
							case "13":
								vote.Votes["lnsupport"] = "yes"
								vote.Votes["sdiffalgo"] = "no"
							case "14":
								fallthrough
							case "15":
								vote.Votes["lnsupport"] = "yes"
								vote.Votes["sdiffalgo"] = "yes"
							}
						case "05":
							vote.Version = "5"
							switch block.RawSTx[i].Vout[1].ScriptPubKey.Hex[4:6] {
							case "00":
								fallthrough
							case "01":
								vote.Votes["lnfeatures"] = "abstain"
							case "02":
								fallthrough
							case "03":
								vote.Votes["lnfeatures"] = "no"
							case "04":
								fallthrough
							case "05":
								vote.Votes["lnfeatures"] = "yes"
							}
						}
					}
					displayBlock.Votes = append(displayBlock.Votes, *vote)
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
// and returns an http.HandlerFunc to use with http.HandleFunc()
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
