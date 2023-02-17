package main

import (
	"context"
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
	var txData *connectors.TxData
	switch os.Args[1] {
	case "ethereum":
		fallthrough
	case "bsc":
		txData, err = connectors.FetchBscTx(cfg, os.Args[2])
	case "polygon":
		txData, err = connectors.FetchPolygonTx(cfg, os.Args[2])
	case "solana":
		txData, err = connectors.FetchSolanaTx(context.TODO(), cfg, os.Args[2])
	default:
		log.Fatalf("unknown chain: %s", os.Args[2])
	}
	if err != nil {
		log.Fatalf("Failed to retrieve tx information: %v", err)
	}

	log.Printf("tx info: sender=%s receiver=%s amount=%s timestamp=%s",
		txData.Source, txData.Destination, txData.Amount, txData.Date)
}
