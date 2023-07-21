package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaReceivedCount            *prometheus.CounterVec
	vaaTotal                    prometheus.Counter
	observationReceivedCount    *prometheus.CounterVec
	observationTotal            prometheus.Counter
	heartbeatReceivedCount      *prometheus.CounterVec
	governorConfigReceivedCount *prometheus.CounterVec
	governorStatusReceivedCount *prometheus.CounterVec
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

	heartbeatReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeat_count_by_guardian",
			Help: "Total number of heartbeat by guardian",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"guardian_node", "type"})

	governorConfigReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "governor_config_count_by_guardian",
			Help: "Total number of governor config by guardian",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"guardian_node", "type"})

	governorStatusReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "governor_status_count_by_guardian",
			Help: "Total number of governor status by guardian",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"guardian_node", "type"})
	return &PrometheusMetrics{
		vaaReceivedCount:            vaaReceivedCount,
		vaaTotal:                    vaaTotal,
		observationReceivedCount:    observationReceivedCount,
		observationTotal:            observationTotal,
		heartbeatReceivedCount:      heartbeatReceivedCount,
		governorConfigReceivedCount: governorConfigReceivedCount,
		governorStatusReceivedCount: governorStatusReceivedCount,
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

// IncObservationWithoutTxHash increases the number of observation without tx hash.
func (m *PrometheusMetrics) IncObservationWithoutTxHash(chain sdk.ChainID) {
	m.observationReceivedCount.WithLabelValues(chain.String(), "without_txhash").Inc()
}

// IncObservationTotal increases the number of observation received from Gossip network.
func (m *PrometheusMetrics) IncObservationTotal() {
	m.observationTotal.Inc()
}

// IncHeartbeatFromGossipNetwork increases the number of heartbeat received by guardian from Gossip network.
func (m *PrometheusMetrics) IncHeartbeatFromGossipNetwork(guardianName string) {
	m.heartbeatReceivedCount.WithLabelValues(guardianName, "gossip").Inc()
}

// IncHeartbeatInserted increases the number of heartbeat inserted in database.
func (m *PrometheusMetrics) IncHeartbeatInserted(guardianName string) {
	m.heartbeatReceivedCount.WithLabelValues(guardianName, "inserted").Inc()
}

// IncGovernorConfigFromGossipNetwork increases the number of guardian config received by guardian from Gossip network.
func (m *PrometheusMetrics) IncGovernorConfigFromGossipNetwork(guardianName string) {
	m.governorConfigReceivedCount.WithLabelValues(guardianName, "gossip").Inc()
}

// IncGovernorConfigInserted increases the number of guardian config inserted in database.
func (m *PrometheusMetrics) IncGovernorConfigInserted(guardianName string) {
	m.governorConfigReceivedCount.WithLabelValues(guardianName, "inserted").Inc()
}

// IncGovernorStatusFromGossipNetwork increases the number of guardian status received by guardian from Gossip network.
func (m *PrometheusMetrics) IncGovernorStatusFromGossipNetwork(guardianName string) {
	m.governorStatusReceivedCount.WithLabelValues(guardianName, "gossip").Inc()
}

// IncGovernorStatusInserted increases the number of guardian status inserted in database.
func (m *PrometheusMetrics) IncGovernorStatusInserted(guardianName string) {
	m.governorStatusReceivedCount.WithLabelValues(guardianName, "inserted").Inc()
}
