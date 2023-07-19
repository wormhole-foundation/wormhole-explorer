package metrics

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct{}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncVaaConsumedQueue is a dummy implementation of IncVaaConsumedQueue.
func (d *DummyMetrics) IncVaaConsumedQueue(chainID uint16) {}

// IncVaaUnfiltered is a dummy implementation of IncVaaUnfiltered.
func (d *DummyMetrics) IncVaaUnfiltered(chainID uint16) {}

// IncOriginTxInserted is a dummy implementation of IncOriginTxInserted.
func (d *DummyMetrics) IncOriginTxInserted(chainID uint16) {}
