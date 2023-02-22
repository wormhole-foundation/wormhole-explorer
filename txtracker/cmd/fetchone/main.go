package main

import (
	"context"
	"log"
	"os"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/connectors"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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

	// get chain ID from args
	chainId, err := vaa.ChainIDFromString(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to convert chain name to chain ID: %v", err)
	}

	// fetch tx data
	txData, err := connectors.FetchTx(context.Background(), cfg, chainId, os.Args[2])
	if err != nil {
		log.Fatalf("Failed to get transaction data: %v", err)
	}

	// print tx details
	log.Printf("tx info: sender=%s receiver=%s amount=%s timestamp=%s",
		txData.Source, txData.Destination, txData.Amount, txData.Timestamp)
}
