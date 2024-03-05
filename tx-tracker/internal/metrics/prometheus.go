package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaTxTrackerCount   *prometheus.CounterVec
	vaaProcesedDuration *prometheus.HistogramVec
	rpcCallCount        *prometheus.CounterVec
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
	vaaProcesedDuration := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "vaa_processed_duration",
		Help: "Duration of vaa processing",
		ConstLabels: map[string]string{
			"environment": environment,
			"service":     serviceName,
		},
		Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60, 120, 300, 600, 1200},
	}, []string{"chain"})

	return &PrometheusMetrics{
		vaaTxTrackerCount:   vaaTxTrackerCount,
		vaaProcesedDuration: vaaProcesedDuration,
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

// AddVaaProcessedDuration adds the duration of vaa processing.
func (m *PrometheusMetrics) AddVaaProcessedDuration(chainID uint16, duration float64) {
	chain := vaa.ChainID(chainID).String()
	m.vaaProcesedDuration.WithLabelValues(chain).Observe(duration)
}

// IncVaaWithoutTxHash increments the number of vaa without tx hash.
func (m *PrometheusMetrics) IncVaaWithoutTxHash(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, "vaa_without_txhash").Inc()
}

// IncVaaWithTxHashFixed increments the number of vaa with tx hash fixed.
func (m *PrometheusMetrics) IncVaaWithTxHashFixed(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, "vaa_txhash_fixed").Inc()
}

// IncCallRpcSuccess increments the number of successful rpc calls.
func (m *PrometheusMetrics) IncCallRpcSuccess(chainID uint16, rpc string) {
	chain := vaa.ChainID(chainID).String()
	m.rpcCallCount.WithLabelValues(chain, rpc, "success").Inc()
}

// IncCallRpcError increments the number of failed rpc calls.
func (m *PrometheusMetrics) IncCallRpcError(chainID uint16, rpc string) {
	chain := vaa.ChainID(chainID).String()
	m.rpcCallCount.WithLabelValues(chain, rpc, "error").Inc()
}
