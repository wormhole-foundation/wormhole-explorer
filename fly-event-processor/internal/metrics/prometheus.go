package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PrometheusMetrics is a Prometheus implementation of Metric interface.
type PrometheusMetrics struct {
	dbLayer             string
	duplicatedVaaCount  *prometheus.CounterVec
	governorStatusCount *prometheus.CounterVec
	governorConfigCount *prometheus.CounterVec
	governorVaaCount    *prometheus.CounterVec
}

// NewPrometheusMetrics returns a new instance of PrometheusMetrics.
func NewPrometheusMetrics(environment string, dbLayer string) *PrometheusMetrics {
	return &PrometheusMetrics{
		duplicatedVaaCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wormscan_fly_event_processor_duplicated_vaa_count",
				Help: "The total number of duplicated VAA processed",
				ConstLabels: map[string]string{
					"environment": environment,
					"service":     serviceName,
				},
			}, []string{"chain", "type", "dblayer"}),
		governorStatusCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wormscan_fly_event_processor_governor_status_count",
				Help: "The total number of governor status processed",
				ConstLabels: map[string]string{
					"environment": environment,
					"service":     serviceName,
				},
			}, []string{"node", "address", "type", "dblayer"}),
		governorConfigCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wormscan_fly_event_processor_governor_config_count",
				Help: "The total number of governor config processed",
				ConstLabels: map[string]string{
					"environment": environment,
					"service":     serviceName,
				},
			}, []string{"node", "address", "type"}),
		governorVaaCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wormscan_fly_event_processor_governor_vaa_count",
				Help: "The total number of governor VAA processed",
				ConstLabels: map[string]string{
					"environment": environment,
					"service":     serviceName,
				},
			}, []string{"chain", "type"}),
	}
}

// IncDuplicatedVaaConsumedQueue increments the total number of duplicated VAA consumed queue.
func (m *PrometheusMetrics) IncDuplicatedVaaConsumedQueue() {
	m.duplicatedVaaCount.WithLabelValues("all", "consumed_queue", m.dbLayer).Inc()
}

// IncDuplicatedVaaProcessed increments the total number of duplicated VAA processed.
func (m *PrometheusMetrics) IncDuplicatedVaaProcessed(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "processed", m.dbLayer).Inc()
}

// IncDuplicatedVaaFailed increments the total number of duplicated VAA failed.
func (m *PrometheusMetrics) IncDuplicatedVaaFailed(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "failed", m.dbLayer).Inc()
}

// IncDuplicatedVaaExpired increments the total number of duplicated VAA expired.
func (m *PrometheusMetrics) IncDuplicatedVaaExpired(chainID sdk.ChainID) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "expired", m.dbLayer).Inc()
}

// IncDuplicatedVaaCanNotFixed increments the total number of duplicated VAA can not fixed.
func (m *PrometheusMetrics) IncDuplicatedVaaCanNotFixed(chainID sdk.ChainID, dbLayer string) {
	chain := chainID.String()
	m.duplicatedVaaCount.WithLabelValues(chain, "can_not_fixed", dbLayer).Inc()
}

// IncGovernorStatusConsumedQueue increments the total number of governor status consumed queue.
func (m *PrometheusMetrics) IncGovernorStatusConsumedQueue() {
	m.governorStatusCount.WithLabelValues("all", "", "consumed_queue", m.dbLayer).Inc()
}

// IncGovernorStatusProcessed increments the total number of governor status processed.
func (m *PrometheusMetrics) IncGovernorStatusProcessed(node string, address string) {
	m.governorStatusCount.WithLabelValues(node, address, "processed", m.dbLayer).Inc()
}

// IncGovernorStatusFailed increments the total number of governor status failed.
func (m *PrometheusMetrics) IncGovernorStatusFailed(node string, address string) {
	m.governorStatusCount.WithLabelValues(node, address, "failed", m.dbLayer).Inc()
}

// IncGovernorStatusUpdateFailed increments the total number of governor status update failed.
func (m *PrometheusMetrics) IncGovernorStatusUpdateFailed(node string, address string, dbLayer string) {
	m.governorStatusCount.WithLabelValues(node, address, "update_failed", dbLayer).Inc()
}

// IncGovernorStatusExpired increments the total number of governor status expired.
func (m *PrometheusMetrics) IncGovernorStatusExpired(node string, address string) {
	m.governorStatusCount.WithLabelValues(node, address, "expired", m.dbLayer).Inc()
}

// IncGovernorConfigConsumedQueue increments the total number of governor config consumed queue.
func (m *PrometheusMetrics) IncGovernorConfigConsumedQueue() {
	m.governorConfigCount.WithLabelValues("all", "", "consumed_queue").Inc()
}

// IncGovernorConfigProcessed increments the total number of governor config processed.
func (m *PrometheusMetrics) IncGovernorConfigProcessed(node string, address string) {
	m.governorConfigCount.WithLabelValues(node, address, "processed").Inc()
}

// IncGovernorConfigFailed increments the total number of governor config failed.
func (m *PrometheusMetrics) IncGovernorConfigFailed(node string, address string) {
	m.governorConfigCount.WithLabelValues(node, address, "failed").Inc()
}

// IncGovernorConfigExpired increments the total number of governor config expired.
func (m *PrometheusMetrics) IncGovernorConfigExpired(node string, address string) {
	m.governorConfigCount.WithLabelValues(node, address, "expired").Inc()
}

// IncGovernorVaaAdded increments the total number of governor VAA added.
func (m *PrometheusMetrics) IncGovernorVaaAdded(chainID sdk.ChainID) {
	chain := chainID.String()
	m.governorVaaCount.WithLabelValues(chain, "added").Inc()
}

// IndGovenorVaaDeleted increments the total number of governor VAA deleted.
func (m *PrometheusMetrics) IndGovenorVaaDeleted(chainID sdk.ChainID) {
	chain := chainID.String()
	m.governorVaaCount.WithLabelValues(chain, "deleted").Inc()
}
