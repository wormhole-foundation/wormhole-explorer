package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct{}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncDuplicatedVaaConsumedQueue dummy implementation.
func (d *DummyMetrics) IncDuplicatedVaaConsumedQueue() {}

// IncDuplicatedVaaProcessed dummy implementation.
func (d *DummyMetrics) IncDuplicatedVaaProcessed(chainID sdk.ChainID) {}

// IncDuplicatedVaaFailed dummy implementation.
func (d *DummyMetrics) IncDuplicatedVaaFailed(chainID sdk.ChainID) {}

// IncDuplicatedVaaExpired dummy implementation.
func (d *DummyMetrics) IncDuplicatedVaaExpired(chainID sdk.ChainID) {}

// IncDuplicatedVaaCanNotFixed dummy implementation.
func (d *DummyMetrics) IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID) {}

// IncGovernorStatusConsumedQueue dummy implementation.
func (d *DummyMetrics) IncGovernorStatusConsumedQueue() {}

// IncGovernorStatusProcessed dummy implementation.
func (d *DummyMetrics) IncGovernorStatusProcessed(node string, address string) {}

// IncGovernorStatusFailed dummy implementation.
func (d *DummyMetrics) IncGovernorStatusFailed(node string, address string) {}

// IncGovernorStatusExpired dummy implementation.
func (d *DummyMetrics) IncGovernorStatusExpired(node string, address string) {}

// IncGovernorVaaAdded dummy implementation.
func (d *DummyMetrics) IncGovernorVaaAdded(chainID sdk.ChainID) {}

// IndGovenorVaaDeleted dummy implementation.
func (d *DummyMetrics) IndGovenorVaaDeleted(chainID sdk.ChainID) {}
