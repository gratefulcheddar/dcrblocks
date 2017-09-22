// Copyright (c) 2014-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
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

func blockHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
	fmt.Fprintf(w, "Input Block Number: %v\n", r.URL.Path[7:])
	t, err := template.ParseFiles("block.html")
	if err != nil {
		log.Fatal(err)
	}
	type Block struct {
		Height string
		Hash   string
		Size   string
	}
	tBlock := Block{"420", "Yes", "Phat"}
	t.Execute(w, tBlock)
}

func getBlock(client *dcrrpcclient.Client, blockStr string) *dcrrpcclient.GetBlockVerboseResult {

	blockInt64, err := strconv.ParseInt(blockStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	blockHash, err := client.GetBlockHash(blockInt64)
	if err != nil {
		log.Fatal(err)
	}
	block, err := client.GetBlockVerbose(blockHash, true)
	if err != nil {
		log.Fatal(err)
	}

	return block
}

func main() {
	// Only override the handlers for notifications you care about.
	// Also note most of these handlers will only be called if you register
	// for notifications.  See the documentation of the dcrrpcclient
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := dcrrpcclient.NotificationHandlers{
		OnBlockConnected: func(blockHeader []byte, transactions [][]byte) {
			log.Printf("Block connected: %v %v", blockHeader, transactions)
		},
		OnBlockDisconnected: func(blockHeader []byte) {
			log.Printf("Block disconnected: %v", blockHeader)
		},
	}

	// Connect to local dcrd RPC server using websockets.
	dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
	certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &dcrrpcclient.ConnConfig{
		Host:         "localhost:9109",
		Endpoint:     "ws",
		User:         "timthomas",
		Pass:         "CfG/BcB1M7q5haQtc6kit34mJycT+bOI",
		Certificates: certs,
	}
	client, err := dcrrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
	log.Println("NotifyBlocks: Registration Complete")

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	getBlock(client, "80193")

	// Start an http server @ localhost:8080 and handle
	// the /block/ URL
	server := http.Server{
		Addr: "localhost:8080",
	}
	http.HandleFunc("/block/", blockHandler)
	server.ListenAndServe()

	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	client.WaitForShutdown()

	// For this example gracefully shutdown the client after 10 seconds.
	// Ordinarily when to shutdown the client is highly application
	// specific.
	log.Println("Client shutdown in 2 seconds...")
	time.AfterFunc(time.Second*2, func() {
		log.Println("Client shutting down...")
		client.Shutdown()
		log.Println("Client shutdown complete.")
	})
}
