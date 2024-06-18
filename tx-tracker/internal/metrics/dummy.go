package metrics

import "time"

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct{}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncVaaConsumedQueue is a dummy implementation of IncVaaConsumedQueue.
func (d *DummyMetrics) IncVaaConsumedQueue(chainID string, source string) {}

// IncVaaUnfiltered is a dummy implementation of IncVaaUnfiltered.
func (d *DummyMetrics) IncVaaUnfiltered(chainID string, source string) {}

// IncOriginTxInserted is a dummy implementation of IncOriginTxInserted.
func (d *DummyMetrics) IncOriginTxInserted(chainID string, source string) {}

// IncDestinationTxInserted is a dummy implementation of IncDestinationTxInserted.
func (d *DummyMetrics) IncDestinationTxInserted(chainID string, source string) {}

// IncVaaWithoutTxHash is a dummy implementation of IncVaaWithoutTxHash.
func (d *DummyMetrics) IncVaaWithoutTxHash(chainID uint16, source string) {}

// IncVaaWithTxHashFixed is a dummy implementation of IncVaaWithTxHashFixed.
func (d *DummyMetrics) IncVaaWithTxHashFixed(chainID uint16, source string) {}

// AddVaaProcessedDuration is a dummy implementation of AddVaaProcessedDuration.
func (d *DummyMetrics) AddVaaProcessedDuration(chainID uint16, duration float64) {}

// IncCallRpcSuccess is a dummy implementation of IncCallRpcSuccess.
func (d *DummyMetrics) IncCallRpcSuccess(chainID uint16, rpc string) {}

// IncCallRpcError is a dummy implementation of IncCallRpcError.
func (d *DummyMetrics) IncCallRpcError(chainID uint16, rpc string) {}

// IncStoreUnprocessedOriginTx is a dummy implementation of IncStoreUnprocessedOriginTx.
func (d *DummyMetrics) IncStoreUnprocessedOriginTx(chainID uint16) {}

// IncVaaProcessed is a dummy implementation of IncVaaProcessed.
func (d *DummyMetrics) IncVaaProcessed(chainID uint16, retry uint8) {}

// IncVaaFailed is a dummy implementation of IncVaaFailed.
func (d *DummyMetrics) IncVaaFailed(chainID uint16, retry uint8) {}

// IncWormchainUnknown is a dummy implementation of IncWormchainUnknown.
func (d *DummyMetrics) IncWormchainUnknown(srcChannel string, dstChannel string) {}

// VaaProcessingDuration increments the duration of VAA processing.
func (m *DummyMetrics) VaaProcessingDuration(chain string, start *time.Time) {}
