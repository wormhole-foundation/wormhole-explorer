package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	measurementCount   *prometheus.CounterVec
	notionalCount      *prometheus.CounterVec
	tokenRequestsCount *prometheus.CounterVec
	processedMessage   *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {

	constLabels := map[string]string{
		"environment": environment,
		"service":     serviceName,
	}

	measurementCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "measurement_count",
			Help:        "Total number of measurement",
			ConstLabels: constLabels,
		}, []string{"measurement", "status"})

	notionalRequestsCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "notional_requests_count_by_symbol",
			Help:        "Total number requests of notional by symbol",
			ConstLabels: constLabels,
		},
		[]string{"symbol", "status"},
	)

	tokenRequestsCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "token_requests_count",
			Help:        "Total number of missing notional by symbol",
			ConstLabels: constLabels,
		},
		[]string{"chain", "token", "status"},
	)

	processedMessage := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "processed_message",
			Help:        "Total number of processed message",
			ConstLabels: constLabels,
		},
		[]string{"chain", "source", "status"},
	)
	return &PrometheusMetrics{
		measurementCount:   measurementCount,
		notionalCount:      notionalRequestsCount,
		tokenRequestsCount: tokenRequestsCount,
		processedMessage:   processedMessage,
	}
}

func (p *PrometheusMetrics) IncFailedMeasurement(measurement string) {
	p.measurementCount.WithLabelValues(measurement, "failed").Inc()
}

func (p *PrometheusMetrics) IncSuccessfulMeasurement(measurement string) {
	p.measurementCount.WithLabelValues(measurement, "successful").Inc()
}

func (p *PrometheusMetrics) IncMissingNotional(symbol string) {
	p.notionalCount.WithLabelValues(symbol, "missing").Inc()
}

func (p *PrometheusMetrics) IncFoundNotional(symbol string) {
	p.notionalCount.WithLabelValues(symbol, "found").Inc()
}

func (p *PrometheusMetrics) IncMissingToken(chain, token string) {
	p.tokenRequestsCount.WithLabelValues(chain, token, "missing").Inc()
}

func (p *PrometheusMetrics) IncFoundToken(chain, token string) {
	p.tokenRequestsCount.WithLabelValues(chain, token, "found").Inc()
}

func (p *PrometheusMetrics) IncExpiredMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "expired").Inc()
}

func (p *PrometheusMetrics) IncInvalidMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "invalid").Inc()
}

func (p *PrometheusMetrics) IncUnprocessedMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "unprocessed").Inc()
}

func (p *PrometheusMetrics) IncProcessedMessage(chain, source string) {
	p.processedMessage.WithLabelValues(chain, source, "processed").Inc()
}
