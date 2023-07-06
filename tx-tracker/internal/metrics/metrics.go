package metrics

const serviceName = "wormscan-tx-tracker"

type Metrics interface {
	IncVaaConsumedQueue(chainID uint16)
	IncVaaUnexpired(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncOriginTxInserted(chainID uint16)
}
