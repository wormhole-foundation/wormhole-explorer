package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func main() {

	// validate commandline arguments
	if len(os.Args) != 5 {
		log.Fatalf("Usage: ./%s <chain name> <tx hash> <block time> <p2p network>\n", os.Args[0])
	}

	// load config settings
	cfg, err := config.LoadFromEnv[config.RpcProviderSettings]()
	if err != nil {
		log.Fatalf("Failed to load credentials from environment: %v", err)
	}

	// get chain ID from args
	chainId, err := vaa.ChainIDFromString(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to convert chain name to chain ID: %v", err)
	}

	blockTime, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert block time to int64: %v", err)
	}
	timestamp := time.Unix(blockTime, 0)

	// fetch tx data
	chains.Initialize(cfg, nil)
	txDetail, err := chains.FetchTx(context.Background(), cfg, chainId, os.Args[2], &timestamp, os.Args[4])
	if err != nil {
		log.Fatalf("Failed to get transaction data: %v", err)
	}

	// print tx details
	log.Printf("tx detail: %+v", txDetail)
}
