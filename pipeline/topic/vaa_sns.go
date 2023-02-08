package topic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/sns"
	"go.uber.org/zap"
)

// SQS represents a VAA queue in SNS.
type SNS struct {
	producer *sns.Producer
	logger   *zap.Logger
}

// NewVAASNS creates a VAA topic in SNS instances.
func NewVAASNS(producer *sns.Producer, logger *zap.Logger) *SNS {
	s := &SNS{
		producer: producer,
		logger:   logger,
	}
	return s
}

// Publish sends the message to a SNS topic.
func (q *SNS) Publish(ctx context.Context, message *Event) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	groupID := fmt.Sprintf("%d/%s", message.ChainID, message.EmitterAddress)
	return q.producer.SendMessage(ctx, groupID, message.ID, string(body))
}
