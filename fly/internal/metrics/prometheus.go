package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaReceivedCount         *prometheus.CounterVec
	vaaTotal                 prometheus.Counter
	observationReceivedCount *prometheus.CounterVec
	observationTotal         prometheus.Counter
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
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

	vaaTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "vaa_total",
			Help: "Total number of vaa from Gossip network",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		})

	observationReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "observation_count_by_chain",
			Help: "Total number of observation by chain",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain", "type"})

	observationTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "observation_total",
			Help: "Total number of observation from Gossip network",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		})

	return &PrometheusMetrics{
		vaaReceivedCount:         vaaReceivedCount,
		vaaTotal:                 vaaTotal,
		observationReceivedCount: observationReceivedCount,
		observationTotal:         observationTotal,
	}
}

// IncVaaFromGossipNetwork increases the number of vaa received by chain from Gossip network.
func (m *PrometheusMetrics) IncVaaFromGossipNetwork(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "gossip").Inc()
}

// IncVaaUnfiltered increases the number of vaa passing through the local deduplicator.
func (m *PrometheusMetrics) IncVaaUnfiltered(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "unfiltered").Inc()
}

// IncVaaConsumedFromQueue increases the number of vaa consumed from SQS queue with deduplication policy.
func (m *PrometheusMetrics) IncVaaConsumedFromQueue(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "consumed-queue").Inc()
}

// IncVaaInserted increases the number of vaa inserted in database.
func (m *PrometheusMetrics) IncVaaInserted(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "inserted").Inc()
}

// IncVaaTotal increases the number of vaa received from Gossip network.
func (m *PrometheusMetrics) IncVaaTotal() {
	m.vaaTotal.Inc()
}

// IncObservationFromGossipNetwork increases the number of observation received by chain from Gossip network.
func (m *PrometheusMetrics) IncObservationFromGossipNetwork(chain sdk.ChainID) {
	m.observationReceivedCount.WithLabelValues(chain.String(), "gossip").Inc()
}

// IncObservationUnfiltered increases the number of observation not filtered
func (m *PrometheusMetrics) IncObservationUnfiltered(chain sdk.ChainID) {
	m.observationReceivedCount.WithLabelValues(chain.String(), "unfiltered").Inc()
}

// IncObservationInserted increases the number of observation inserted in database.
func (m *PrometheusMetrics) IncObservationInserted(chain sdk.ChainID) {
	m.observationReceivedCount.WithLabelValues(chain.String(), "inserted").Inc()
}

// IncObservationTotal increases the number of observation received from Gossip network.
func (m *PrometheusMetrics) IncObservationTotal() {
	m.observationTotal.Inc()
}
