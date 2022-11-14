package queue

import (
	"context"
	"encoding/base64"
	"fly/internal/sqs"
	"fmt"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type SQS struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan *Message
	logger   *zap.Logger
}

func NewVAASQS(producer *sqs.Producer, consumer *sqs.Consumer, logger *zap.Logger) *SQS {
	return &SQS{
		producer: producer,
		consumer: consumer,
		ch:       make(chan *Message, 1000),
		logger:   logger}
}

func (q *SQS) Publish(_ context.Context, v *vaa.VAA, data []byte) error {
	body := base64.StdEncoding.EncodeToString(data)
	groupID := fmt.Sprintf("%d/%s", v.EmitterChain, v.EmitterAddress)
	return q.producer.SendMessage(groupID, v.MessageID(), body)
}

func (q *SQS) Consume() <-chan *Message {
	go func() {
		for {
			messages, err := q.consumer.GetMessages()
			if err != nil {
				q.logger.Error("Error getting messages from SQS", zap.Error(err))
				continue
			}
			for _, msg := range messages {
				body, err := base64.StdEncoding.DecodeString(*msg.Body)
				if err != nil {
					q.logger.Error("Error decoding message from SQS", zap.Error(err))
					continue
				}
				q.ch <- &Message{
					Data: body,
					Ack: func() {
						if err := q.consumer.DeleteMessage(msg); err != nil {
							q.logger.Error("Error deleting message from SQS", zap.Error(err))
						}
					},
				}
			}

		}
	}()
	return q.ch
}

func (q *SQS) Close() {
	close(q.ch)
}
