package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

const serviceName = "wormscan-fly-event-processor"

type Metrics interface {
	IncDuplicatedVaaConsumedQueue()
	IncDuplicatedVaaProcessed(chainID sdk.ChainID)
	IncDuplicatedVaaFailed(chainID sdk.ChainID)
	IncDuplicatedVaaExpired(chainID sdk.ChainID)
	IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID)
	IncGovernorStatusConsumedQueue()
	IncGovernorStatusProcessed(node string, address string)
	IncGovernorStatusFailed(node string, address string)
	IncGovernorStatusExpired(node string, address string)
	IncGovernorVaaAdded(chainID sdk.ChainID)
	IndGovenorVaaDeleted(chainID sdk.ChainID)
}

// IncDuplicatedVaaConsumedQueue increments the counter of consumed queue
type IncConsumedQueue func()

/*
// ProcessorFunc is a function to process a governor message.
type ProcessorFunc func(context.Context, *Params) error
*/
