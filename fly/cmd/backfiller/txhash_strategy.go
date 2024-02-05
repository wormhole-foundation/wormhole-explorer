package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"github.com/wormhole-foundation/wormhole-explorer/fly/txhash"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func workerTxHash(ctx context.Context, repo *storage.Repository, txHashStore txhash.TxHashStore, line string) error {
	tokens := strings.Split(line, ",")
	if len(tokens) != 4 {
		return fmt.Errorf("invalid line: %s", line)
	}

	intChainID, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing chain id: %v\n", err)
	}

	//remove 0x from txhash
	if len(tokens[3]) < 3 {
		return fmt.Errorf("invalid txhash: %s", tokens[3])
	}
	// if token starts with 0x remove it
	if tokens[3][:2] == "0x" {
		tokens[3] = tokens[3][2:]
	}

	txHash := strings.ToLower(tokens[3])

	id := fmt.Sprintf("%d/%s%s", intChainID, tokens[1], tokens[2])

	vaaTxHash := txhash.TxHash{
		ChainID:  vaa.ChainID(intChainID),
		Emitter:  tokens[1],
		Sequence: tokens[2],
		TxHash:   txHash,
	}

	err = txHashStore.Set(ctx, id, vaaTxHash)
	if err != nil {
		return fmt.Errorf("error upserting vaa: %v\n", err)
	}

	return nil

}
