package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	duplicatedVaaCount *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string) *PrometheusMetrics {
	return &PrometheusMetrics{
		duplicatedVaaCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wormscan_fly_event_processor_duplicated_vaa_count",
				Help: "The total number of duplicated VAA processed",
				ConstLabels: map[string]string{
					"environment": environment,
					"service":     serviceName,
				},
			}, []string{"chain", "type"}),
	}
}

func (m *PrometheusMetrics) IncDuplicatedVaaConsumedQueue() {
	m.duplicatedVaaCount.WithLabelValues("all", "consumed_queue").Inc()
}

func (m *PrometheusMetrics) IncDuplicatedVaaProcessed(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "processed").Inc()
}

func (m *PrometheusMetrics) IncDuplicatedVaaFailed(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "failed").Inc()
}

func (m *PrometheusMetrics) IncDuplicatedVaaExpired(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "expired").Inc()
}

func (m *PrometheusMetrics) IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "can_not_fixed").Inc()
}
