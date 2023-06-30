package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type PrometheusMetrics struct {
	destinationTrxCount *prometheus.CounterVec
	lastBlock           *prometheus.GaugeVec
	currentBlock        *prometheus.GaugeVec
	requestsTotal       *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {

	constLabels := map[string]string{
		"environment": environment,
		"service":     serviceName,
	}

	destinationTrxCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "destination_tx_count_by_chain",
			Help:        "Total number of destination trx by chain",
			ConstLabels: constLabels,
		}, []string{"chain"})

	lastBlock := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "last_block_by_chain",
			Help:        "Last block number by chain",
			ConstLabels: constLabels,
		}, []string{"chain"})

	currentBlock := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "current_block_by_chain",
			Help:        "Current block number by chain",
			ConstLabels: constLabels,
		}, []string{"chain"})

	requestsTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "rpc_requests_total_by_chain",
			Help:        "Count all http requests by status code, method and path.",
			ConstLabels: constLabels,
		},
		[]string{"client", "operation", "status_code"},
	)

	return &PrometheusMetrics{
		lastBlock:           lastBlock,
		currentBlock:        currentBlock,
		destinationTrxCount: destinationTrxCount,
		requestsTotal:       requestsTotal,
	}
}

func (m *PrometheusMetrics) SetLastBlock(chain sdk.ChainID, block uint64) {
	m.lastBlock.WithLabelValues(chain.String()).Set(float64(block))
}

func (m *PrometheusMetrics) SetCurrentBlock(chain sdk.ChainID, block uint64) {
	m.currentBlock.WithLabelValues(chain.String()).Set(float64(block))
}

func (m *PrometheusMetrics) IncDestinationTrxSaved(chain sdk.ChainID) {
	m.destinationTrxCount.WithLabelValues(chain.String()).Inc()
}

func (m *PrometheusMetrics) IncRpcRequest(client string, operation string, statusCode int) {
	m.requestsTotal.WithLabelValues(client, operation, strconv.Itoa(statusCode)).Inc()
}
