package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

const serviceName = "wormscan-pipeline"

// Metrics is a metrics interface.
type Metrics interface {
	IncVaaFromMongoStream(chainID uint16)
	IncVaaSendNotification(chainID uint16)

	IncVaaWithoutTxHash(chainID uint16)
	IncVaaWithTxHashFixed(chainID uint16)

	IncVaaSendNotificationFromGossipSQS(chainID sdk.ChainID)
	IncVaaFromGossipSQS(chainID uint16)
	IncVaaFailedProcessing(chainID sdk.ChainID, reason string)
}
