package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"go.uber.org/zap"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan *VaaEvent
	chSize   int
	logger   *zap.Logger
}

// NewVAASQS creates a VAA queue in SQS instances.
func NewVAASQS(producer *sqs.Producer, consumer *sqs.Consumer, logger *zap.Logger, opts ...SQSOption) *SQS {
	s := &SQS{
		producer: producer,
		consumer: consumer,
		chSize:   10,
		logger:   logger}
	for _, opt := range opts {
		opt(s)
	}
	s.ch = make(chan *VaaEvent, s.chSize)
	return s
}

// WithChannelSize allows to specify an channel size when setting a value.
func WithChannelSize(size int) SQSOption {
	return func(d *SQS) {
		d.chSize = size
	}
}

// Publish sends the message to a SQS queue.
func (q *SQS) Publish(_ context.Context, message *VaaEvent) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	groupID := fmt.Sprintf("%d/%s", message.ChainID, message.EmitterAddress)
	deduplicationID := fmt.Sprintf("%d/%s/%d", message.ChainID, message.EmitterAddress, message.Sequence)
	return q.producer.SendMessage(groupID, deduplicationID, string(body))
}

// Close closes all consumer resources.
func (q *SQS) Close() {
	close(q.ch)
}
