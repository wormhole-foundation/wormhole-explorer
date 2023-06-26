package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

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

// IncObservationTotal increases the number of observation received from Gossip network.
func (d *DummyMetrics) IncObservationTotal() {}

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
