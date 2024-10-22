package topic

import (
	"context"
	"encoding/json"

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
func (s *SNS) Publish(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	s.logger.Debug("Publishing message", zap.String("groupID", event.Digest))
	err = s.producer.SendMessage(ctx, event.ChainID, event.Digest, event.Digest, string(body))
	if err == nil {
		s.metrics.IncVaaSendNotification(uint16(event.ChainID))
	} else {
		// Alert error pushing event.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"vaaID":     event.ID,
				"groupID":   event.Digest,
				"messageID": event.Digest,
			},
			Error: err,
		}
		s.alertClient.CreateAndSend(ctx, pipelineAlert.ErrorPushEventSNS, alertContext)
	}
	return err
}
