package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaReceivedCount              *prometheus.CounterVec
	vaaTotal                      prometheus.Counter
	observationReceivedCount      *prometheus.CounterVec
	observationTotal              prometheus.Counter
	observationReceivedByGuardian *prometheus.CounterVec
	heartbeatReceivedCount        *prometheus.CounterVec
	governorConfigReceivedCount   *prometheus.CounterVec
	governorStatusReceivedCount   *prometheus.CounterVec
	maxSequenceCacheCount         *prometheus.CounterVec
	txHashSearchCount             *prometheus.CounterVec
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

	observationReceivedByGuardian := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "observation_count_by_guardian",
			Help: "Total number of observation by guardian",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"guardian_address", "type"})

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
	maxSequenceCacheCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "max_sequence_cache_count_by_chain",
			Help: "Total number of errors when updating max sequence cache",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain"})
	txHashSearchCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tx_hash_search_count_by_store",
			Help: "Total number of errors when updating max sequence cache",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"store", "action"})
	return &PrometheusMetrics{
		vaaReceivedCount:              vaaReceivedCount,
		vaaTotal:                      vaaTotal,
		observationReceivedCount:      observationReceivedCount,
		observationTotal:              observationTotal,
		heartbeatReceivedCount:        heartbeatReceivedCount,
		governorConfigReceivedCount:   governorConfigReceivedCount,
		governorStatusReceivedCount:   governorStatusReceivedCount,
		maxSequenceCacheCount:         maxSequenceCacheCount,
		txHashSearchCount:             txHashSearchCount,
		observationReceivedByGuardian: observationReceivedByGuardian,
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

// IncVaaSendNotification increases the number of vaa send notifcations to pipeline.
func (m *PrometheusMetrics) IncVaaSendNotification(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "send-notification").Inc()
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

// IncObservationInvalidGuardian increases the number of invalid guardian in observation from Gossip network.
func (m *PrometheusMetrics) IncObservationInvalidGuardian(address string) {
	m.observationReceivedByGuardian.WithLabelValues(address, "invalid_guardian").Inc()
}

// IncObservationInvalidGuardian increases the number of bad signer in observation from Gossip network.
func (m *PrometheusMetrics) IncObservationBadSigner(address string) {
	m.observationReceivedByGuardian.WithLabelValues(address, "bad_signer").Inc()
}

// IncObservationInvalidGuardian increases the number of bad signer in observation from Gossip network.
func (m *PrometheusMetrics) IncObservationValid(address string) {
	m.observationReceivedByGuardian.WithLabelValues(address, "valid").Inc()
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

// IncMaxSequenceCacheError increases the number of errors when updating max sequence cache.
func (m *PrometheusMetrics) IncMaxSequenceCacheError(chain sdk.ChainID) {
	m.maxSequenceCacheCount.WithLabelValues(chain.String()).Inc()
}

func (m *PrometheusMetrics) IncFoundTxHash(t string) {
	m.txHashSearchCount.WithLabelValues(t, "found").Inc()
}

func (m *PrometheusMetrics) IncNotFoundTxHash(t string) {
	m.txHashSearchCount.WithLabelValues(t, "not_found").Inc()
}
