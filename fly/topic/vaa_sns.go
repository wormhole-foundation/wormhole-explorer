package topic

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/sns"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"go.uber.org/zap"
)

// SNSProducer is a producer for SNS.
type SNSProducer struct {
	producer    *sns.Producer
	alertClient alert.AlertClient
	metrics     metrics.Metrics
	logger      *zap.Logger
}

// NewSNSProducer creates a new SNSProducer.
func NewSNSProducer(producer *sns.Producer, alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger) *SNSProducer {
	return &SNSProducer{
		producer:    producer,
		alertClient: alertClient,
		metrics:     metrics,
		logger:      logger,
	}
}

// Push pushes a VAAEvent to SNS.
func (p *SNSProducer) Push(ctx context.Context, event *NotificationEvent) error {
	return nil
}
