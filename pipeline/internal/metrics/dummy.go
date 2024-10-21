package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct {
}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncVaaFromMongoStream increments the vaa received count from mongo stream.
func (m *DummyMetrics) IncVaaFromMongoStream(chainID uint16) {}

// IncVaaSendNotification increments the vaa received count send notification.
func (m *DummyMetrics) IncVaaSendNotification(chainID uint16) {}

// IncVaaWithoutTxHash increments the vaa received count without tx hash.
func (m *DummyMetrics) IncVaaWithoutTxHash(chainID uint16) {}

// IncVaaWithTxHashFixed increments the vaa received count with tx hash fixed.
func (m *DummyMetrics) IncVaaWithTxHashFixed(chainID uint16) {}

func (m *DummyMetrics) IncVaaSendNotificationFromGossipSQS(chainID sdk.ChainID)   {}
func (m *DummyMetrics) IncVaaFromGossipSQS(chainID uint16)                        {}
func (m *DummyMetrics) IncVaaFailedProcessing(chainID sdk.ChainID, reason string) {}
