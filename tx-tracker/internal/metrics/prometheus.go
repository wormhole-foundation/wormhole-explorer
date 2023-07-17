package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaTxTrackerCount *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	vaaTxTrackerCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vaa_tx_tracker_count_by_chain",
			Help: "Total number of vaa processed by tx tracker by chain",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain", "type"})
	return &PrometheusMetrics{
		vaaTxTrackerCount: vaaTxTrackerCount,
	}
}

// IncVaaConsumedQueue increments the number of consumed VAA.
func (m *PrometheusMetrics) IncVaaConsumedQueue(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, "consumed_queue").Inc()
}

// IncVaaUnfiltered increments the number of unfiltered VAA.
func (m *PrometheusMetrics) IncVaaUnfiltered(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, "unfiltered").Inc()
}

// IncOriginTxInserted increments the number of inserted origin tx.
func (m *PrometheusMetrics) IncOriginTxInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, "origin_tx_inserted").Inc()
}
