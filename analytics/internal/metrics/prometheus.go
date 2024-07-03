package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	measurementCount      *prometheus.CounterVec
	notionalCount         *prometheus.CounterVec
	tokenRequestsCount    *prometheus.CounterVec
	processedMessage      *prometheus.CounterVec
	vaaProcessingDuration *prometheus.HistogramVec
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
		[]string{"chain", "source", "status", "retry"},
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
		measurementCount:      measurementCount,
		notionalCount:         notionalRequestsCount,
		tokenRequestsCount:    tokenRequestsCount,
		processedMessage:      processedMessage,
		vaaProcessingDuration: vaaProcessingDuration,
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

func (p *PrometheusMetrics) IncExpiredMessage(chain, source string, retry uint8) {
	p.processedMessage.WithLabelValues(chain, source, "expired", fmt.Sprintf("%d", retry)).Inc()
}

func (p *PrometheusMetrics) IncInvalidMessage(chain, source string, retry uint8) {
	p.processedMessage.WithLabelValues(chain, source, "invalid", fmt.Sprintf("%d", retry)).Inc()
}

func (p *PrometheusMetrics) IncUnprocessedMessage(chain, source string, retry uint8) {
	p.processedMessage.WithLabelValues(chain, source, "unprocessed", fmt.Sprintf("%d", retry)).Inc()
}

func (p *PrometheusMetrics) IncProcessedMessage(chain, source string, retry uint8) {
	p.processedMessage.WithLabelValues(chain, source, "processed", fmt.Sprintf("%d", retry)).Inc()
}

func (p *PrometheusMetrics) VaaProcessingDuration(chain string, start *time.Time) {
	if start == nil {
		return
	}
	elapsed := float64(time.Since(*start).Nanoseconds()) / 1e9
	p.vaaProcessingDuration.WithLabelValues(chain).Observe(elapsed)
}
