package metrics

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct {
}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncVaaFromGossipNetwork increases the number of vaa received by chain from Gossip network.
func (d *DummyMetrics) IncVaaFromGossipNetwork(chain sdk.ChainID) {}

// IncVaaUnfiltered increases the number of vaa passing through the local deduplicator.
func (d *DummyMetrics) IncVaaUnfiltered(chain sdk.ChainID) {}

// IncVaaConsumedFromQueue increases the number of vaa consumed from SQS queue with deduplication policy.
func (d *DummyMetrics) IncVaaConsumedFromQueue(chain sdk.ChainID) {}

// IncVaaInserted increases the number of vaa inserted into the database.
func (d *DummyMetrics) IncVaaInserted(chain sdk.ChainID) {}

// IncVaaTotal increases the number of vaa received from Gossip network.
func (d *DummyMetrics) IncVaaTotal() {}

// IncObservationFromGossipNetwork increases the number of observation received by chain from Gossip network.
func (d *DummyMetrics) IncObservationFromGossipNetwork(chain sdk.ChainID) {}

// IncObservationUnfiltered increases the number of observation not filtered
func (d *DummyMetrics) IncObservationUnfiltered(chain sdk.ChainID) {}

// IncObservationInserted increases the number of observation inserted in database.
func (d *DummyMetrics) IncObservationInserted(chain sdk.ChainID) {}

// IncObservationWithoutTxHash increases the number of observation without tx hash.
func (d *DummyMetrics) IncObservationWithoutTxHash(chain sdk.ChainID) {}

// IncVaaSendNotification increases the number of vaa send notifcations to pipeline.
func (d *DummyMetrics) IncVaaSendNotification(chain sdk.ChainID) {}

// IncObservationTotal increases the number of observation received from Gossip network.
func (d *DummyMetrics) IncObservationTotal() {}

// IncBatchObservationTotal increases the number of batch observation messages received from Gossip network.
func (d *DummyMetrics) IncBatchObservationTotal(batchSize uint) {}

// IncObservationInvalidGuardian increases the number of invalid guardian in observation from Gossip network.
func (m *DummyMetrics) IncObservationInvalidGuardian(address string) {}

// IncObservationInvalidGuardian increases the number of bad signer in observation from Gossip network.
func (m *DummyMetrics) IncObservationBadSigner(address string) {}

// IncObservationInvalidGuardian increases the number of bad signer in observation from Gossip network.
func (m *DummyMetrics) IncObservationValid(address string) {}

// IncHeartbeatFromGossipNetwork increases the number of heartbeat received by guardian from Gossip network.
func (d *DummyMetrics) IncHeartbeatFromGossipNetwork(guardianName string) {}

// IncHeartbeatInserted increases the number of heartbeat inserted in database.
func (d *DummyMetrics) IncHeartbeatInserted(guardianName string) {}

// IncGovernorConfigFromGossipNetwork increases the number of guardian config received by guardian from Gossip network.
func (d *DummyMetrics) IncGovernorConfigFromGossipNetwork(guardianName string) {}

// IncGovernorConfigInserted increases the number of guardian config inserted in database.
func (d *DummyMetrics) IncGovernorConfigInserted(guardianName string) {}

// IncGovernorStatusFromGossipNetwork increases the number of guardian status received by guardian from Gossip network.
func (d *DummyMetrics) IncGovernorStatusFromGossipNetwork(guardianName string) {}

// IncGovernorStatusInserted increases the number of guardian status inserted in database.
func (d *DummyMetrics) IncGovernorStatusInserted(guardianName string) {}

// IncMaxSequenceCacheError increases the number of errors when updating max sequence cache.
func (d *DummyMetrics) IncMaxSequenceCacheError(chain sdk.ChainID) {}

func (m *DummyMetrics) IncFoundTxHash(t string) {}

func (m *DummyMetrics) IncNotFoundTxHash(t string) {}

func (m *DummyMetrics) IncConsistencyLevelByChainID(chainID sdk.ChainID, consistenceLevel uint8) {}

func (m *DummyMetrics) IncDuplicateVaaByChainID(chain sdk.ChainID) {}

func (m *DummyMetrics) VaaProcessingDuration(chain sdk.ChainID, start *time.Time) {}
