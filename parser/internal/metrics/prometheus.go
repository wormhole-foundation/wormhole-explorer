package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	dbLayer                       string
	vaaParseCount                 *prometheus.CounterVec
	vaaPayloadParserRequest       *prometheus.CounterVec
	vaaPayloadParserResponseCount *prometheus.CounterVec
	processedMessage              *prometheus.CounterVec
	vaaProcessingDuration         *prometheus.HistogramVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string, dbLayer string) *PrometheusMetrics {
	constLabels := map[string]string{
		"environment": environment,
		"service":     serviceName,
	}
	vaaParseCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "parse_vaa_count_by_chain",
			Help:        "Total number of vaa parser by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "type", "dbLayer"})
	vaaPayloadParserRequestCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "parse_vaa_payload_request_count_by_chain",
			Help:        "Total number of request to payload parser component by chain",
			ConstLabels: constLabels,
		}, []string{"chain"})
	vaaPayloadParserResponseCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "parse_vaa_payload_response_count_by_chain",
			Help:        "Total number of response from payload parser component by chain",
			ConstLabels: constLabels,
		}, []string{"chain", "status"})
	processedMessage := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "processed_message",
			Help:        "Total number of processed message",
			ConstLabels: constLabels,
		},
		[]string{"chain", "source", "status"},
	)
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
		vaaParseCount:                 vaaParseCount,
		vaaPayloadParserRequest:       vaaPayloadParserRequestCount,
		vaaPayloadParserResponseCount: vaaPayloadParserResponseCount,
		processedMessage:              processedMessage,
		vaaProcessingDuration:         vaaProcessingDuration,
	}
}

// IncVaaConsumedQueue increments the number of consumed VAA.
func (m *PrometheusMetrics) IncVaaConsumedQueue(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "consumed", m.dbLayer).Inc()
}

// IncVaaUnfiltered increments the number of unfiltered VAA.
func (m *PrometheusMetrics) IncVaaUnfiltered(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "unfiltered", m.dbLayer).Inc()
}

// IncVaaUnexpired increments the number of unexpired VAA.
func (m *PrometheusMetrics) IncVaaUnexpired(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "unexpired", m.dbLayer).Inc()
}

// IncVaaParsed increments the number of parsed VAA.
func (m *PrometheusMetrics) IncVaaParsed(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "parsed", m.dbLayer).Inc()
}

// IncVaaParsedInserted increments the number of parsed VAA inserted into database.
func (m *PrometheusMetrics) IncVaaParsedInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "inserted", m.dbLayer).Inc()
}

// IncVaaAttestationPropertiesInserted increment the number of attestation properties inserted into database.
func (m *PrometheusMetrics) IncVaaAttestationPropertiesInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "attestation_properties_inserted").Inc()
}

// IncParseVaaInserted increments the number of parsed VAA inserted into database.
func (m *PrometheusMetrics) IncParseVaaInserted(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaParseCount.WithLabelValues(chain, "parsed_vaa_inserted").Inc()
}

// IncVaaPayloadParserRequestCount increments the number of vaa payload parser request.
func (m *PrometheusMetrics) IncVaaPayloadParserRequestCount(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaPayloadParserRequest.WithLabelValues(chain).Inc()
}

// IncVaaPayloadParserErrorCount increments the number of vaa payload parser error.
func (m *PrometheusMetrics) IncVaaPayloadParserErrorCount(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaPayloadParserResponseCount.WithLabelValues(chain, "failed").Inc()
}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser success.
func (m *PrometheusMetrics) IncVaaPayloadParserSuccessCount(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaPayloadParserResponseCount.WithLabelValues(chain, "success").Inc()
}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser not found.
func (m *PrometheusMetrics) IncVaaPayloadParserNotFoundCount(chainID uint16) {
	chain := vaa.ChainID(chainID).String()
	m.vaaPayloadParserResponseCount.WithLabelValues(chain, "not_found").Inc()
}

// IncExpiredMessage increments the number of expired message.
func (p *PrometheusMetrics) IncExpiredMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "expired").Inc()
}

// IncUnprocessedMessage increments the number of unprocessed message.
func (p *PrometheusMetrics) IncUnprocessedMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "unprocessed").Inc()
}

// IncProcessedMessage increments the number of processed message.
func (p *PrometheusMetrics) IncProcessedMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "processed").Inc()
}

// VaaProcessingDuration increases the duration of vaa processing.
func (p *PrometheusMetrics) VaaProcessingDuration(chain string, start *time.Time) {
	if start == nil {
		return
	}
	elapsed := float64(time.Since(*start).Nanoseconds()) / 1e9
	p.vaaProcessingDuration.WithLabelValues(chain).Observe(elapsed)
}
