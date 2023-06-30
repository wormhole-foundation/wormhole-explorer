package metrics

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

// NoopMetrics is a no-op implementation of the Metrics interface.
type NoopMetrics struct {
}

// NewNoopMetrics returns a new instance of NoopMetrics.
func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

func (m *NoopMetrics) SetLastBlock(chain sdk.ChainID, block uint64) {
}

func (m *NoopMetrics) SetCurrentBlock(chain sdk.ChainID, block uint64) {
}

func (m *NoopMetrics) IncDestinationTrxSaved(chain sdk.ChainID) {
}

func (m *NoopMetrics) IncRpcRequest(client string, method string, statusCode int) {
}
