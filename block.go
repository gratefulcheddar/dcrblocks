package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrrpcclient"
)

type displayBlock struct {
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

	Transactions    []regular
	Votes           []vote
	TicketPurchases []ticketPurchase
	Revocations     []revocation
}

type regular struct {
	Amount float64
	TxID   string
}

type vote struct {
	TxID       string
	VoteResult *parsedVote
}

type parsedVote struct {
	Version string
	Votes   map[string]string
}

type ticketPurchase struct {
	TxID     string
	Maturity string
}

type revocation struct {
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

			displayBlock := parseBlock(block, blockSubsidy, currentBlockHeight)

			// Load the template with the block data
			t.Execute(w, displayBlock)
		}
	}
}

// parseBlock parses a GetBlockVerboseResult into a displayBlock for use
// with the block.html templates
func parseBlock(block *dcrjson.GetBlockVerboseResult, blockSubsidy *dcrjson.GetBlockSubsidyResult, maxHeight int64) *displayBlock {

	displayBlock := new(displayBlock)
	displayBlock.BlockHeight = block.Height
	displayBlock.BlockSize = block.Size
	displayBlock.BlockVersion = block.Version
	displayBlock.Confirmations = block.Confirmations
	displayBlock.Hash = block.Hash
	displayBlock.MerkleRoot = block.MerkleRoot
	if displayBlock.BlockHeight != maxHeight {
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
	// regularTransactions with the minimum information
	// required to display to user.
	for i := 0; i < len(block.RawTx); i++ {
		regularTransaction := new(regular)
		regularTransaction.TxID = block.RawTx[i].Txid
		for _, value := range block.RawTx[i].Vin {
			regularTransaction.Amount += value.AmountIn
		}
		if block.Height == 0 {
			regularTransaction.Amount = 0
		}
		displayBlock.Transactions = append(displayBlock.Transactions, *regularTransaction)
	}

	// Loop through the Raw Stake Transactions, creating two
	// slices, one of ticketPurchaseTransactions and one of
	// voteTransactions, with the minimum information required
	// to display to user.
	for i := 0; i < len(block.RawSTx); i++ {
		if block.RawSTx[i].Vout[0].ScriptPubKey.Type == "stakesubmission" {
			ticketPurchase := new(ticketPurchase)
			ticketPurchase.TxID = block.RawSTx[i].Txid

			if block.RawSTx[i].Confirmations > 256 {
				ticketPurchase.Maturity = "Mature"
			} else {
				ticketPurchase.Maturity = "Immature"
			}

			displayBlock.TicketPurchases = append(displayBlock.TicketPurchases, *ticketPurchase)

		} else if block.RawSTx[i].Vout[0].ScriptPubKey.Type == "stakerevoke" {
			revocation := new(revocation)
			revocation.TxID = block.RawSTx[i].Txid
			displayBlock.Revocations = append(displayBlock.Revocations, *revocation)
		} else {
			vote := new(vote)
			vote.TxID = block.RawSTx[i].Txid
			// Parse Vote - TODO: Make this automatic
			if len(block.RawSTx[i].Vout[1].ScriptPubKey.Hex) > 8 {
				vote.VoteResult = parseVoteScript(block.RawSTx[i].Vout[1].ScriptPubKey.Hex)
			}
			displayBlock.Votes = append(displayBlock.Votes, *vote)
		}
	}
	return displayBlock
}

func parseVoteScript(scriptPubKeyHex string) *parsedVote {
	voteResults := new(parsedVote)

	voteResults.Votes = make(map[string]string)
	switch scriptPubKeyHex[8:10] {
	case "04":
		voteResults.Version = "4"
		switch scriptPubKeyHex[4:6] {
		case "00":
			fallthrough
		case "01":
			voteResults.Votes["lnsupport"] = "abstain"
			voteResults.Votes["sdiffalgo"] = "abstain"
		case "02":
			fallthrough
		case "03":
			voteResults.Votes["lnsupport"] = "abstain"
			voteResults.Votes["sdiffalgo"] = "no"
		case "04":
			fallthrough
		case "05":
			voteResults.Votes["lnsupport"] = "abstain"
			voteResults.Votes["sdiffalgo"] = "yes"
		case "08":
			fallthrough
		case "09":
			voteResults.Votes["lnsupport"] = "no"
			voteResults.Votes["sdiffalgo"] = "abstain"
		case "0a":
			fallthrough
		case "0b":
			voteResults.Votes["lnsupport"] = "no"
			voteResults.Votes["sdiffalgo"] = "no"
		case "0c":
			fallthrough
		case "0d":
			voteResults.Votes["lnsupport"] = "no"
			voteResults.Votes["sdiffalgo"] = "yes"
		case "10":
			fallthrough
		case "11":
			voteResults.Votes["lnsupport"] = "yes"
			voteResults.Votes["sdiffalgo"] = "abstain"
		case "12":
			fallthrough
		case "13":
			voteResults.Votes["lnsupport"] = "yes"
			voteResults.Votes["sdiffalgo"] = "no"
		case "14":
			fallthrough
		case "15":
			voteResults.Votes["lnsupport"] = "yes"
			voteResults.Votes["sdiffalgo"] = "yes"
		}
	case "05":
		voteResults.Version = "5"
		switch scriptPubKeyHex[4:6] {
		case "00":
			fallthrough
		case "01":
			voteResults.Votes["lnfeatures"] = "abstain"
		case "02":
			fallthrough
		case "03":
			voteResults.Votes["lnfeatures"] = "no"
		case "04":
			fallthrough
		case "05":
			voteResults.Votes["lnfeatures"] = "yes"
		}
	}
	return voteResults
}
