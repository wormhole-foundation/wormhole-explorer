package metrics

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	return &PrometheusMetrics{}
}
