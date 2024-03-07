package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	expiredCacheResponseCount *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	vaaTxTrackerCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "expired_cache_response",
			Help: "Total expired cache response by key",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"key"})

	return &PrometheusMetrics{
		expiredCacheResponseCount: vaaTxTrackerCount,
	}
}

func (m *PrometheusMetrics) IncExpiredCacheResponse(key string) {
	m.expiredCacheResponseCount.WithLabelValues(key).Inc()
}

type noOpMetrics struct{}

func (s *noOpMetrics) IncExpiredCacheResponse(_ string) {
}

func NewNoOpMetrics() Metrics {
	return &noOpMetrics{}
}
