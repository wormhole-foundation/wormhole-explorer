package main

import (
	"log"
	"os"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/connectors"
)

func main() {

	// validate commandline arguments
	if len(os.Args) != 3 {
		log.Fatalf("Usage: ./%s <chain name> <tx hash>\n", os.Args[0])
	}

	// load config settings
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load credentials from environment: %v", err)
	}

	// query data for the specified chain
	switch os.Args[1] {
	case "ethereum":
		txInfo, err := connectors.FetchEthereumTx(cfg, os.Args[2])
		if err != nil {
			log.Fatalf("Failed to retrieve tx information: %v", err)
		}

		log.Printf("tx info: %+v", txInfo)
	default:
		log.Fatalf("unknown chain: %s", os.Args[2])
	}

}
