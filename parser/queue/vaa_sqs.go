package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"go.uber.org/zap"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan *ConsumerMessage
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
	s.ch = make(chan *ConsumerMessage, s.chSize)
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

// Consume returns the channel with the received messages from SQS queue.
func (q *SQS) Consume(ctx context.Context) <-chan *ConsumerMessage {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				messages, err := q.consumer.GetMessages()
				if err != nil {
					q.logger.Error("Error getting messages from SQS", zap.Error(err))
					continue
				}
				expiredAt := time.Now().Add(q.consumer.GetVisibilityTimeout())
				for _, msg := range messages {
					var body VaaEvent
					err := json.Unmarshal([]byte(*msg.Body), &body)
					if err != nil {
						q.logger.Error("Error decoding message from SQS", zap.Error(err))
						continue
					}
					q.ch <- &ConsumerMessage{
						Data: &body,
						Ack: func() {
							if err := q.consumer.DeleteMessage(msg); err != nil {
								q.logger.Error("Error deleting message from SQS", zap.Error(err))
							}
						},
						IsExpired: func() bool {
							return expiredAt.Before(time.Now())
						},
					}
				}
			}
		}
	}()
	return q.ch
}

// Close closes all consumer resources.
func (q *SQS) Close() {
	close(q.ch)
}
