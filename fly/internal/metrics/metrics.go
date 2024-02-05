package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

const serviceName = "wormscan-fly"

type Metrics interface {
	// vaa metrics
	IncVaaFromGossipNetwork(chain sdk.ChainID)
	IncVaaUnfiltered(chain sdk.ChainID)
	IncVaaConsumedFromQueue(chain sdk.ChainID)
	IncVaaInserted(chain sdk.ChainID)
	IncVaaSendNotification(chain sdk.ChainID)
	IncVaaTotal()

	// observation metrics
	IncObservationFromGossipNetwork(chain sdk.ChainID)
	IncObservationUnfiltered(chain sdk.ChainID)
	IncObservationInserted(chain sdk.ChainID)
	IncObservationWithoutTxHash(chain sdk.ChainID)
	IncObservationTotal()

	// heartbeat metrics
	IncHeartbeatFromGossipNetwork(guardianName string)
	IncHeartbeatInserted(guardianName string)

	// governor config metrics
	IncGovernorConfigFromGossipNetwork(guardianName string)
	IncGovernorConfigInserted(guardianName string)

	// governor status metrics
	IncGovernorStatusFromGossipNetwork(guardianName string)
	IncGovernorStatusInserted(guardianName string)

	// max sequence cache metrics
	IncMaxSequenceCacheError(chain sdk.ChainID)

	// tx hash metrics
	IncFoundTxHash(t string)
	IncNotFoundTxHash(t string)
}
