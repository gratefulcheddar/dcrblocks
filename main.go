// Copyright (c) 2014-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrutil"
)

// makeHandler creates a new rpcclient, passes it to the
// input function for executing, shuts down the client,
// and returns an http.HandlerFunc to use with http.HandleFunc()
func makeHandler(fn func(http.ResponseWriter, *http.Request, *rpcclient.Client)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Connect to local dcrd RPC server using websockets.
		dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
		certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
		if err != nil {
			log.Fatal(err)
		}
		connCfg := &rpcclient.ConnConfig{
			Host:         "localhost:9109",
			Endpoint:     "ws",
			User:         "cheesepool",
			Pass:         "dcrblocks",
			Certificates: certs,
		}
		client, err := rpcclient.New(connCfg, nil)
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
	http.HandleFunc("/transaction/", makeHandler(handleTransaction))
	http.HandleFunc("/address/", makeHandler(handleAddress))
	http.HandleFunc("/", makeHandler(handleIndex))
	log.Printf("Server started. Address = %v", server.Addr)
	server.ListenAndServe()

}
