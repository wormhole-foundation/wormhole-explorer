package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a metrics implementation for Prometheus.
type PrometheusMetrics struct {
	vaaReceivedCount *prometheus.CounterVec
	vaaTxHashCount   *prometheus.CounterVec
}

// NewPrometheusMetrics creates a new PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	vaaReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vaa_count_by_chain",
			Help: "Total number of vaa by chain",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain", "type"})

	vaaTxHashCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vaa_txhash_count_by_chain",
			Help: "Total number of vaa by chain",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain", "type"})

	return &PrometheusMetrics{
		vaaReceivedCount: vaaReceivedCount,
		vaaTxHashCount:   vaaTxHashCount,
	}
}

// IncVaaFromMongoStream increments the vaa received count from mongo stream.
func (m *PrometheusMetrics) IncVaaFromMongoStream(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaReceivedCount.WithLabelValues(chain, "mongo-stream").Inc()
}

// IncVaaSendNotification increments the vaa received count send event to SNS.
func (m *PrometheusMetrics) IncVaaSendNotification(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaReceivedCount.WithLabelValues(chain, "publish-notification").Inc()
}

// IncVaaWithoutTxHash increments the vaa received count without tx hash.
func (m *PrometheusMetrics) IncVaaWithoutTxHash(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxHashCount.WithLabelValues(chain, "vaa-without-txhash").Inc()
}

// IncVaaWithTxHashFixed increments the vaa received count with tx hash fixed.
func (m *PrometheusMetrics) IncVaaWithTxHashFixed(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxHashCount.WithLabelValues(chain, "vaa-with-txhash-fixed").Inc()
}
