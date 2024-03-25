package metrics

const serviceName = "wormscan-tx-tracker"

type Metrics interface {
	IncVaaConsumedQueue(chainID string, source string)
	IncVaaUnfiltered(chainID string, source string)
	IncOriginTxInserted(chainID string, source string)
	IncVaaWithoutTxHash(chainID uint16)
	IncVaaWithTxHashFixed(chainID uint16)
	IncDestinationTxInserted(chainID string, source string)
	AddVaaProcessedDuration(chainID uint16, duration float64)
	IncCallRpcSuccess(chainID uint16, rpc string)
	IncCallRpcError(chainID uint16, rpc string)
	IncStoreUnprocessedOriginTx(chainID uint16)
	IncVaaProcessed(chainID uint16, retry uint8)
	IncVaaFailed(chainID uint16, retry uint8)
}
