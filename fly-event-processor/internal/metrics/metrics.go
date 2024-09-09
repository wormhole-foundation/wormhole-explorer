package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

const serviceName = "wormscan-fly-event-processor"

type Metrics interface {
	IncDuplicatedVaaConsumedQueue()
	IncDuplicatedVaaProcessed(chainID sdk.ChainID)
	IncDuplicatedVaaFailed(chainID sdk.ChainID)
	IncDuplicatedVaaExpired(chainID sdk.ChainID)
	// TODO: remove dbLayer after db migration.
	IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID, dbLayer string)
	IncGovernorStatusConsumedQueue()
	IncGovernorStatusProcessed(node string, address string)
	IncGovernorStatusFailed(node string, address string)
	// TODO: remove metrics after db migration.
	IncGovernorStatusUpdateFailed(node string, address string, dbLayer string)
	IncGovernorStatusExpired(node string, address string)
	IncGovernorConfigConsumedQueue()
	IncGovernorConfigProcessed(node string, address string)
	IncGovernorConfigFailed(node string, address string)
	IncGovernorConfigExpired(node string, address string)
	IncGovernorVaaAdded(chainID sdk.ChainID)
	IndGovenorVaaDeleted(chainID sdk.ChainID)
}

// IncDuplicatedVaaConsumedQueue increments the counter of consumed queue
type IncConsumedQueue func()
