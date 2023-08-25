package topic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/sns"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	deduplicationID := fmt.Sprintf("gossip-event-%s", event.Payload.ID)
	p.logger.Debug("Publishing signedVaa event", zap.String("groupID", event.Payload.ID))
	err = p.producer.SendMessage(ctx, event.Payload.ID, deduplicationID, string(body))
	if err == nil {
		p.metrics.IncVaaSendNotification(vaa.ChainID(event.Payload.EmitterChain))
	}
	return err
}
