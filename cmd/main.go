package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/infoslack/blockchain"
)

func main() {
	serverPort := flag.String("port", "5000", "http port number where server will run")
	flag.Parse()

	blockchain := myblockchain.NewBlockchain()
	nodeID := strings.Replace(myblockchain.UUID(), "-", "", -1)

	log.Printf("Starting blockchain HTTP Server. Listening at port %q", *serverPort)

	http.Handle("/", myblockchain.NewHandler(blockchain, nodeID))
	http.ListenAndServe(fmt.Sprintf(":%s", *serverPort), nil)
}
