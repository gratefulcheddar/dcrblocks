# dcrblocks
golang block explorer for decred

This is an incomplete project I made from scratch to help myself learn Golang, web programming, and Decred's blockchain. I followed Sau Sheong Chang's _Go Web Programming_ for most of the design. 

dcrblocks requires a running instance of dcrd with address indexing enabled (`addrindex=1`), an RPC username of `cheesepool`, and an RPC password of `dcrblocks` (can be changed in `main.go`). All of it's querying of dcrd is done by first parsing the URL. It handles `/block/<block number>`, `/transaction/<transaction id>`, and `/address/<decred address>` requests. Index loads the last 10 blocks and requires a manual refresh for updates. It uses port 8080.

