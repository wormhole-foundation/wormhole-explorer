package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Metrics struct {
	vaaReceivedCount *prometheus.CounterVec
	vaaTotal         prometheus.Counter
}

const serviceName = "wormscan-fly"

func NewMetrics(environment string) *Metrics {

	vaaReceivedCount := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vaa_count_by_chain",
			Help: "Total number of vaa by chain",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		}, []string{"chain", "type"})

	vaaTotal := promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "vaa_total",
			Help: "Total number of vaa from Gossip network",
			ConstLabels: map[string]string{
				"environment": environment,
				"service":     serviceName,
			},
		})

	return &Metrics{vaaReceivedCount: vaaReceivedCount, vaaTotal: vaaTotal}
}

// IncVaaFromGossipNetwork increases the number of vaa received by chain from Gossip network.
func (m *Metrics) IncVaaFromGossipNetwork(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "gossip").Inc()
}

// IncVaaUnfiltered increases the number of vaa passing through the local deduplicator.
func (m *Metrics) IncVaaUnfiltered(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "unfiltered").Inc()
}

// IncVaaConsumedFromQueue increases the number of vaa consumed from SQS queue with deduplication policy.
func (m *Metrics) IncVaaConsumedFromQueue(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "consumed-queue").Inc()
}

// IncVaaInserted increases the number of vaa inserted in database.
func (m *Metrics) IncVaaInserted(chain sdk.ChainID) {
	m.vaaReceivedCount.WithLabelValues(chain.String(), "inserted").Inc()
}

// IncVaaTotal increases the number of vaa received from Gossip network.
func (m *Metrics) IncVaaTotal() {
	m.vaaTotal.Inc()
}
