package metrics

import "time"

// NoopMetrics is a no-op implementation of the Metrics interface.
type NoopMetrics struct {
}

// NewNoopMetrics returns a new instance of NoopMetrics.
func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

func (p *NoopMetrics) IncFailedMeasurement(measurement string) {
}

func (p *NoopMetrics) IncSuccessfulMeasurement(measurement string) {
}

func (p *NoopMetrics) IncMissingNotional(symbol string) {
}

func (p *NoopMetrics) IncFoundNotional(symbol string) {
}

func (p *NoopMetrics) IncMissingToken(chain, token string) {
}

func (p *NoopMetrics) IncFoundToken(chain, token string) {
}

func (p *NoopMetrics) IncExpiredMessage(chain, source string, retry uint8) {
}

func (p *NoopMetrics) IncInvalidMessage(chain, source string, retry uint8) {
}

func (p *NoopMetrics) IncUnprocessedMessage(chain, source string, retry uint8) {
}

func (p *NoopMetrics) IncProcessedMessage(chain, source string, retry uint8) {
}

func (m *NoopMetrics) VaaProcessingDuration(chain string, start *time.Time) {
}
