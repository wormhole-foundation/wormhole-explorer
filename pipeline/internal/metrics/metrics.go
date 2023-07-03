package metrics

const serviceName = "wormscan-pipeline"

// Metrics is a metrics interface.
type Metrics interface {
	IncVaaFromMongoStream(chainID uint16)
	IncVaaSendNotification(chainID uint16)

	IncVaaWithoutTxHash(chainID uint16)
	IncVaaWithTxHashFixed(chainID uint16)
}
