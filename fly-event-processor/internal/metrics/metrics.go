package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

const serviceName = "wormscan-fly-event-processor"

type Metrics interface {
	IncDuplicatedVaaConsumedQueue()
	IncDuplicatedVaaProcessed(chainID sdk.ChainID)
	IncDuplicatedVaaFailed(chainID sdk.ChainID)
	IncDuplicatedVaaExpired(chainID sdk.ChainID)
	IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID)
}
