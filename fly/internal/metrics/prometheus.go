package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaReceivedCount              *prometheus.CounterVec
	vaaTotal                      prometheus.Counter
	observationReceivedCount      *prometheus.CounterVec
	observationTotal              prometheus.Counter
	batchObservationTotal         prometheus.Counter
	batchSizeObservations         prometheus.Gauge
	observationReceivedByGuardian *prometheus.CounterVec
	heartbeatReceivedCount        *prometheus.CounterVec
	governorConfigReceivedCount   *prometheus.CounterVec
	governorStatusReceivedCount   *prometheus.CounterVec
	maxSequenceCacheCount         *prometheus.CounterVec
	txHashSearchCount             *prometheus.CounterVec
	consistenceLevelChainCount    *prometheus.CounterVec
	duplicateVaaByChainCount      *prometheus.CounterVec
	vaaProcessingDuration         *prometheus.HistogramVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string, dbLayer string) *PrometheusMetrics {
	service := serviceName
	if dbLayer == config.DbLayerPostgres {
		service = fmt.Sprintf("%s-%s", serviceName, dbLayer)
	}
	constLabels := map[string]string{
		"environment": environment,
		"service":     service,
	}
	vaaReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "vaa_count_by_chain",
			Help:        "Total number of vaa by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "type"})

	vaaTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "vaa_total",
			Help:        "Total number of vaa from Gossip network",
			ConstLabels: constLabels,
		})

	observationReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "observation_count_by_chain",
			Help:        "Total number of observation by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "type"})

	observationTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "observation_total",
			Help:        "Total number of observation from Gossip network",
			ConstLabels: constLabels,
		})

	batchObservationTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name:        "batch_observation_total",
			Help:        "Total number of batch observation messages from Gossip network",
			ConstLabels: constLabels,
		})

	batchSizeObservations := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name:        "batch_size_observations",
			Help:        "Batch-observation sizes incoming from Gossip network",
			ConstLabels: constLabels,
		})

	observationReceivedByGuardian := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "observation_count_by_guardian",
			Help:        "Total number of observation by guardian",
			ConstLabels: constLabels,
		}, []string{"guardian_address", "type"})

	heartbeatReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "heartbeat_count_by_guardian",
			Help:        "Total number of heartbeat by guardian",
			ConstLabels: constLabels,
		}, []string{"guardian_node", "type"})

	governorConfigReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "governor_config_count_by_guardian",
			Help:        "Total number of governor config by guardian",
			ConstLabels: constLabels,
		}, []string{"guardian_node", "type"})

	governorStatusReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "governor_status_count_by_guardian",
			Help:        "Total number of governor status by guardian",
			ConstLabels: constLabels,
		}, []string{"guardian_node", "type"})
	maxSequenceCacheCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "max_sequence_cache_count_by_chain",
			Help:        "Total number of errors when updating max sequence cache",
			ConstLabels: constLabels,
		}, []string{"chain"})
	txHashSearchCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "tx_hash_search_count_by_store",
			Help:        "Total number of errors when updating max sequence cache",
			ConstLabels: constLabels,
		}, []string{"store", "action"})
	consistenceLevelChainCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "consistence_level_count_by_chain",
			Help:        "Total number of consistence level by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "consistence_level"})
	duplicateVaaByChainCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "duplicate_vaa_count_by_chain",
			Help:        "Total number of duplicate vaa by chain",
			ConstLabels: constLabels,
		}, []string{"chain"})
	vaaProcessingDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "vaa_processing_duration_seconds",
			Help:        "Duration of all vaa processing by chain.",
			ConstLabels: constLabels,
			Buckets:     []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60, 120, 300, 600, 1200},
		},
		[]string{"chain"},
	)
	return &PrometheusMetrics{
		vaaReceivedCount:              vaaReceivedCount,
		vaaTotal:                      vaaTotal,
		observationReceivedCount:      observationReceivedCount,
		observationTotal:              observationTotal,
		batchObservationTotal:         batchObservationTotal,
		batchSizeObservations:         batchSizeObservations,
		heartbeatReceivedCount:        heartbeatReceivedCount,
		governorConfigReceivedCount:   governorConfigReceivedCount,
		governorStatusReceivedCount:   governorStatusReceivedCount,
		maxSequenceCacheCount:         maxSequenceCacheCount,
		txHashSearchCount:             txHashSearchCount,
		observationReceivedByGuardian: observationReceivedByGuardian,
		consistenceLevelChainCount:    consistenceLevelChainCount,
		duplicateVaaByChainCount:      duplicateVaaByChainCount,
		vaaProcessingDuration:         vaaProcessingDuration,
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

// IncBatchObservationTotal increases the number of batch observation messages received from Gossip network.
func (m *PrometheusMetrics) IncBatchObservationTotal(batchSize uint) {
	m.batchObservationTotal.Inc()
	m.batchSizeObservations.Add(float64(batchSize))
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

// IncConsistencyLevelByChainID increases the number of errors when updating max sequence cache.
func (m *PrometheusMetrics) IncConsistencyLevelByChainID(chainID sdk.ChainID, consistenceLevel uint8) {
	m.consistenceLevelChainCount.WithLabelValues(chainID.String(), fmt.Sprintf("%d", consistenceLevel)).Inc()
}

// IncDuplicateVaaByChainID increases the number of duplicate vaa by chain.
func (m *PrometheusMetrics) IncDuplicateVaaByChainID(chain sdk.ChainID) {
	m.duplicateVaaByChainCount.WithLabelValues(chain.String()).Inc()
}

// VaaProcessingDuration increases the duration of vaa processing.
func (m *PrometheusMetrics) VaaProcessingDuration(chain sdk.ChainID, start *time.Time) {
	if start == nil {
		return
	}
	elapsed := float64(time.Since(*start).Nanoseconds()) / 1e9
	m.vaaProcessingDuration.WithLabelValues(chain.String()).Observe(elapsed)
}
