package watcher

import "context"

// ContractTracker is an interface for tracking contracts
// It Tracks contract operations and persist the tx data
// BackfillContract is used to backfill the contract data from the past
type ContractWatcher interface {
	Start(ctx context.Context) error
	Close()
}
