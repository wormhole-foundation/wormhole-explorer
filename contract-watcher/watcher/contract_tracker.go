package watcher

import "context"

// ContractTracker is an interface for tracking contracts
// It Tracks contract operations and persist the tx data
// Backfill is used to backfill the contract data from the past
type ContractWatcher interface {
	Start(ctx context.Context) error
	Close()
	Backfill(ctx context.Context, fromBlock uint64, toBlock uint64, pageSize uint64, persistBlock bool)
}
