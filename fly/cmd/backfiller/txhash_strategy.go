package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func workerTxHash(ctx context.Context, repo *storage.Repository, line string) error {
	tokens := strings.Split(line, ",")
	if len(tokens) != 4 {
		return fmt.Errorf("invalid line: %s", line)
	}

	intChainID, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing chain id: %v\n", err)
	}

	now := time.Now()

	vaaTxHash := storage.VaaIdTxHashUpdate{
		ChainID:   vaa.ChainID(intChainID),
		Emitter:   tokens[1],
		Sequence:  tokens[2],
		TxHash:    tokens[3],
		UpdatedAt: &now,
	}

	err = repo.UpsertTxHash(ctx, vaaTxHash)
	if err != nil {
		return fmt.Errorf("error upserting vaa: %v\n", err)
	}

	return nil

}
