package metrics

const serviceName = "wormscan-tx-tracker-v2"

type Metrics interface {
	IncVaaConsumedQueue(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncOriginTxInserted(chainID uint16)
}
