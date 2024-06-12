package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	expiredCacheResponseCount *prometheus.CounterVec
	originRequestsCount       *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	constLabels := map[string]string{
		"environment": environment,
		"service":     serviceName,
	}

	vaaTxTrackerCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "expired_cache_response",
			Help:        "Total expired cache response by key",
			ConstLabels: constLabels,
		}, []string{"key"})

	originRequestsCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_origin_requests_total",
			Help:        "Count all http requests by origin, method and path.",
			ConstLabels: constLabels,
		},
		[]string{"origin"},
	)

	return &PrometheusMetrics{
		expiredCacheResponseCount: vaaTxTrackerCount,
		originRequestsCount:       originRequestsCount,
	}
}

func (m *PrometheusMetrics) IncExpiredCacheResponse(key string) {
	m.expiredCacheResponseCount.WithLabelValues(key).Inc()
}

func (m *PrometheusMetrics) IncOrigin(origin string) {
	m.originRequestsCount.WithLabelValues(origin).Inc()
}

type noOpMetrics struct{}

func (s *noOpMetrics) IncExpiredCacheResponse(_ string) {}

func (s *noOpMetrics) IncOrigin(_ string) {}

func NewNoOpMetrics() Metrics {
	return &noOpMetrics{}
}
