package queue

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type SQSOption func(*SQS)

type SQS struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan *Message
	chSize   int
	logger   *zap.Logger
}

func NewVAASQS(producer *sqs.Producer, consumer *sqs.Consumer, logger *zap.Logger, opts ...SQSOption) *SQS {
	s := &SQS{
		producer: producer,
		consumer: consumer,
		chSize:   10,
		logger:   logger}
	for _, opt := range opts {
		opt(s)
	}
	s.ch = make(chan *Message, s.chSize)
	return s
}

func WithChannelSize(size int) SQSOption {
	return func(d *SQS) {
		d.chSize = size
	}
}

func (q *SQS) Publish(_ context.Context, v *vaa.VAA, data []byte) error {
	body := base64.StdEncoding.EncodeToString(data)
	groupID := fmt.Sprintf("%d/%s", v.EmitterChain, v.EmitterAddress)
	return q.producer.SendMessage(groupID, v.MessageID(), body)
}

func (q *SQS) Consume(ctx context.Context) <-chan *Message {
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
					body, err := base64.StdEncoding.DecodeString(*msg.Body)
					if err != nil {
						q.logger.Error("Error decoding message from SQS", zap.Error(err))
						continue
					}
					//TODO check if callback is better than channel
					q.ch <- &Message{
						Data: body,
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

func (q *SQS) Close() {
	close(q.ch)
}
