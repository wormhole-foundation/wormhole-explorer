package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	vaaTxTrackerCount        *prometheus.CounterVec
	vaaProcesedDuration      *prometheus.HistogramVec
	rpcCallCount             *prometheus.CounterVec
	storeUnprocessedOriginTx *prometheus.CounterVec
	vaaProcessed             *prometheus.CounterVec
	wormchainUnknown         *prometheus.CounterVec
	vaaProcessingDuration    *prometheus.HistogramVec
	globalTxCount            *prometheus.CounterVec
	operationTxCount         *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	constLabels := map[string]string{
		"environment": environment,
		"service":     serviceName,
	}
	vaaTxTrackerCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "vaa_tx_tracker_count_by_chain",
			Help:        "Total number of vaa processed by tx tracker by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "source", "type"})
	vaaProcesedDuration := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "vaa_processed_duration",
		Help:        "Duration of vaa processing",
		ConstLabels: constLabels,
		Buckets:     []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60, 120, 300, 600, 1200},
	}, []string{"chain"})
	rpcCallCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "rpc_call_count_by_chain",
			Help:        "Total number of rpc calls by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "rpc", "status"})
	storeUnprocessedOriginTx := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "store_unprocessed_origin_tx",
			Help:        "Total number of unprocessed origin tx",
			ConstLabels: constLabels,
		}, []string{"chain"})
	vaaProcessed := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "vaa_processed",
			Help:        "Total number of processed vaa with retry context",
			ConstLabels: constLabels,
		}, []string{"chain", "retry", "status"})
	wormchainUnknown := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "wormchain_unknown",
			Help:        "Total number of unknown wormchain",
			ConstLabels: constLabels,
		}, []string{"srcChannel", "dstChannel"})
	vaaProcessingDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "vaa_processing_duration_seconds",
			Help:        "Duration of all vaa processing by chain.",
			ConstLabels: constLabels,
			Buckets:     []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 30, 60, 120, 300, 600, 1200},
		},
		[]string{"chain"},
	)
	globalTxCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "global_tx_count_by_chain",
			Help:        "Total number of global tx processed by tx tracker by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "type"})
	operationTxCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "operation_tx_count_by_chain",
			Help:        "Total number of operation tx processed by tx tracker by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "type"})
	return &PrometheusMetrics{
		vaaTxTrackerCount:        vaaTxTrackerCount,
		vaaProcesedDuration:      vaaProcesedDuration,
		rpcCallCount:             rpcCallCount,
		storeUnprocessedOriginTx: storeUnprocessedOriginTx,
		vaaProcessed:             vaaProcessed,
		wormchainUnknown:         wormchainUnknown,
		vaaProcessingDuration:    vaaProcessingDuration,
		globalTxCount:            globalTxCount,
		operationTxCount:         operationTxCount,
	}
}

// IncVaaConsumedQueue increments the number of consumed VAA.
func (m *PrometheusMetrics) IncVaaConsumedQueue(chainID string, source string) {
	m.vaaTxTrackerCount.WithLabelValues(chainID, source, "consumed_queue").Inc()
}

// IncVaaUnfiltered increments the number of unfiltered VAA.
func (m *PrometheusMetrics) IncVaaUnfiltered(chainID string, source string) {
	m.vaaTxTrackerCount.WithLabelValues(chainID, source, "unfiltered").Inc()
}

// IncOriginTxInserted increments the number of inserted origin tx.
func (m *PrometheusMetrics) IncOriginTxInserted(chainID string, source string) {
	m.vaaTxTrackerCount.WithLabelValues(chainID, source, "origin_tx_inserted").Inc()
}

// IncDestinationTxInserted increments the number of inserted destination tx.
func (m *PrometheusMetrics) IncDestinationTxInserted(chainID string, source string) {
	m.vaaTxTrackerCount.WithLabelValues(chainID, source, "destination_tx_inserted").Inc()
}

// AddVaaProcessedDuration adds the duration of vaa processing.
func (m *PrometheusMetrics) AddVaaProcessedDuration(chainID uint16, duration float64) {
	chain := vaa.ChainID(chainID).String()
	m.vaaProcesedDuration.WithLabelValues(chain).Observe(duration)
}

// IncVaaWithoutTxHash increments the number of vaa without tx hash.
func (m *PrometheusMetrics) IncVaaWithoutTxHash(chainID uint16, source string) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, source, "vaa_without_txhash").Inc()
}

// IncVaaWithTxHashFixed increments the number of vaa with tx hash fixed.
func (m *PrometheusMetrics) IncVaaWithTxHashFixed(chainID uint16, source string) {
	chain := vaa.ChainID(chainID).String()
	m.vaaTxTrackerCount.WithLabelValues(chain, source, "vaa_txhash_fixed").Inc()
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

// IncStoreUnprocessedOriginTx increments the number of unprocessed origin tx.
func (m *PrometheusMetrics) IncStoreUnprocessedOriginTx(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.storeUnprocessedOriginTx.WithLabelValues(chain).Inc()
}

// IncVaaProcessed increments the number of processed vaa.
func (m *PrometheusMetrics) IncVaaProcessed(chainID uint16, retry uint8) {
	chain := vaa.ChainID(chainID).String()
	m.vaaProcessed.WithLabelValues(chain, strconv.Itoa(int(retry)), "success").Inc()
}

// IncVaaFailed increments the number of failed vaa.
func (m *PrometheusMetrics) IncVaaFailed(chainID uint16, retry uint8) {
	chain := vaa.ChainID(chainID).String()
	m.vaaProcessed.WithLabelValues(chain, strconv.Itoa(int(retry)), "failed").Inc()
}

// IncWormchainUnknown increments the number of unknown wormchain.
func (m *PrometheusMetrics) IncWormchainUnknown(srcChannel string, dstChannel string) {
	m.wormchainUnknown.WithLabelValues(srcChannel, dstChannel).Inc()
}

// VaaProcessingDuration increases the duration of vaa processing.
func (p *PrometheusMetrics) VaaProcessingDuration(chain string, start *time.Time) {
	if start == nil {
		return
	}
	elapsed := float64(time.Since(*start).Nanoseconds()) / 1e9
	p.vaaProcessingDuration.WithLabelValues(chain).Observe(elapsed)
}

// IncOperationTxSourceInserted increments the number of inserted operation tx.
func (p *PrometheusMetrics) IncOperationTxSourceInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	p.operationTxCount.WithLabelValues(chain, "source").Inc()
}

// IncGlobalTxSourceInserted increments the number of inserted global tx.
func (p *PrometheusMetrics) IncGlobalTxSourceInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	p.globalTxCount.WithLabelValues(chain, "source").Inc()
}

// IncOperationTxTargetInserted increments the number of inserted operation tx.
func (p *PrometheusMetrics) IncOperationTxTargetInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	p.operationTxCount.WithLabelValues(chain, "target").Inc()
}

// IncGlobalTxDestinationTxInserted increments the number of inserted global tx.
func (p *PrometheusMetrics) IncGlobalTxDestinationTxInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	p.globalTxCount.WithLabelValues(chain, "target").Inc()
}
