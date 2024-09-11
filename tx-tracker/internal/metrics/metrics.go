package metrics

import "time"

const serviceName = "wormscan-tx-tracker"

type Metrics interface {
	IncVaaConsumedQueue(chainID string, source string)
	IncVaaUnfiltered(chainID string, source string)
	IncOriginTxInserted(chainID string, source string)
	IncVaaWithoutTxHash(chainID uint16, source string)
	IncVaaWithTxHashFixed(chainID uint16, source string)
	IncDestinationTxInserted(chainID string, source string)
	AddVaaProcessedDuration(chainID uint16, duration float64)
	IncCallRpcSuccess(chainID uint16, rpc string)
	IncCallRpcError(chainID uint16, rpc string)
	IncStoreUnprocessedOriginTx(chainID uint16)
	IncVaaProcessed(chainID uint16, retry uint8)
	IncVaaFailed(chainID uint16, retry uint8)
	IncWormchainUnknown(srcChannel string, dstChannel string)
	VaaProcessingDuration(chain string, start *time.Time)
	// TODO: remove after database migration.
	IncOperationTxSourceInserted(chainID uint16)
	// TODO: remove after database migration.
	IncGlobalTxSourceInserted(chainID uint16)
	// TODO: remove after database migration.
	IncOperationTxTargetInserted(chainID uint16)
	// TODO: remove after database migration.
	IncGlobalTxDestinationTxInserted(chainID uint16)
}
