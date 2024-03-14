package metrics

const serviceName = "wormscan-tx-tracker"

type Metrics interface {
	IncVaaConsumedQueue(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncOriginTxInserted(chainID uint16)
	IncVaaWithoutTxHash(chainID uint16)
	IncVaaWithTxHashFixed(chainID uint16)
	AddVaaProcessedDuration(chainID uint16, duration float64)
	IncCallRpcSuccess(chainID uint16, rpc string)
	IncCallRpcError(chainID uint16, rpc string)
	IncStoreUnprocessedOriginTx(chainID uint16)
	IncVaaProcessed(chainID uint16, retry uint8)
	IncVaaFailed(chainID uint16, retry uint8)
}
