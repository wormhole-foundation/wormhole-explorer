package watcher

import (
	"testing"
	"time"

	"github.com/test-go/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/solana"
)

func Test_waitForSolanaBlock(t *testing.T) {
	block := uint64(233566448)
	lastestBlock := &solana.GetLatestBlockResult{
		Block:     233566448,
		Timestamp: time.Now(),
	}
	waitForSolanaBlock(block, lastestBlock)
	assert.Equal(t, time.Since(lastestBlock.Timestamp) > 20*time.Second, true)
}

func Test_noWaitForSolanaBlock(t *testing.T) {
	block := uint64(233566248)
	lastestBlock := &solana.GetLatestBlockResult{
		Block:     233566448,
		Timestamp: time.Now(),
	}
	waitForSolanaBlock(block, lastestBlock)
	assert.Equal(t, time.Since(lastestBlock.Timestamp) < 1*time.Second, true)
}
