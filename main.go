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

	"github.com/decred/dcrrpcclient"
	"github.com/decred/dcrutil"
)

func handleBlock(w http.ResponseWriter, r *http.Request, client *dcrrpcclient.Client) {

	blockStr := r.URL.Path[7:]

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

	t, err := template.ParseFiles("block.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, block)

	client.Shutdown()
}

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
			User:         "timthomas",
			Pass:         "CfG/BcB1M7q5haQtc6kit34mJycT+bOI",
			Certificates: certs,
		}
		client, err := dcrrpcclient.New(connCfg, nil)
		if err != nil {
			log.Fatal(err)
		}
		fn(w, r, client)
	}
}

func main() {

	server := http.Server{
		Addr: "localhost:8080",
	}
	http.HandleFunc("/block/", makeHandler(handleBlock))
	log.Printf("Server started. Address = %v", server.Addr)
	server.ListenAndServe()

}
