package topic

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	pipelineAlert "github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/sns"
	"go.uber.org/zap"
)

// SQS represents a VAA queue in SNS.
type SNS struct {
	producer    *sns.Producer
	alertClient alert.AlertClient
	metrics     metrics.Metrics
	logger      *zap.Logger
}

// NewVAASNS creates a VAA topic in SNS instances.
func NewVAASNS(producer *sns.Producer, alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger) *SNS {
	s := &SNS{
		producer:    producer,
		alertClient: alertClient,
		metrics:     metrics,
		logger:      logger,
	}
	return s
}

// Publish sends the message to a SNS topic.
func (s *SNS) Publish(ctx context.Context, message SnsMessage) error {
	body, err := message.Body()
	if err != nil {
		return err
	}

	s.logger.Debug("Publishing message", zap.String("groupID", message.GetGroupID()))
	err = s.producer.SendMessage(ctx, message.GetChainID(), message.GetGroupID(), message.GetDeduplicationID(), string(body))
	if err == nil {
		s.metrics.IncVaaSendNotification(uint16(message.GetChainID()))
	} else {
		// Alert error pushing event.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"groupID":   message.GetGroupID(),
				"messageID": message.GetDeduplicationID(),
			},
			Error: err,
		}
		s.alertClient.CreateAndSend(ctx, pipelineAlert.ErrorPushEventSNS, alertContext)
	}
	return err
}
