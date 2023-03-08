package watcher

// ContractTracker is an interface for tracking contracts
// It Tracks contract operations and persist the tx data
// BackfillContract is used to backfill the contract data from the past
type ContractWatcher interface {
	WatchContract() error
	BackfillContract(fromBlock int64, toBlock int64) error
}
