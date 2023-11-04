package producer

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
func (p *SNSProducer) Push(ctx context.Context, n *Notification) error {
	body, err := json.Marshal(n.Event)
	if err != nil {
		return err
	}
	deduplicationID := fmt.Sprintf("gossip-event-%s", n.ID)
	p.logger.Debug("Publishing signedVaa event", zap.String("groupID", n.ID))
	err = p.producer.SendMessage(ctx, n.ID, deduplicationID, string(body))
	if err == nil {
		p.metrics.IncVaaSendNotification(vaa.ChainID(n.EmitterChain))
	}
	return err
}
