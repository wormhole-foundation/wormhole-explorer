package queue

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan Message
	chSize   int
	wg       sync.WaitGroup
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
	s.ch = make(chan Message, s.chSize)
	return s
}

// WithChannelSize allows to specify an channel size when setting a value.
func WithChannelSize(size int) SQSOption {
	return func(d *SQS) {
		d.chSize = size
	}
}

// Publish sends the message to a SQS queue.
func (q *SQS) Publish(_ context.Context, v *vaa.VAA, data []byte) error {
	body := base64.StdEncoding.EncodeToString(data)
	groupID := fmt.Sprintf("%d/%s", v.EmitterChain, v.EmitterAddress)
	return q.producer.SendMessage(groupID, v.MessageID(), body)
}

// Consume returns the channel with the received messages from SQS queue.
func (q *SQS) Consume(ctx context.Context) <-chan Message {
	go func() {
		for {
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
				q.wg.Add(1)
				q.ch <- &sqsConsumerMessage{
					id:        msg.ReceiptHandle,
					data:      body,
					wg:        &q.wg,
					logger:    q.logger,
					consumer:  q.consumer,
					expiredAt: expiredAt,
					ctx:       ctx,
				}
			}
			q.wg.Wait()
		}
	}()
	return q.ch
}

// Close closes all consumer resources.
func (q *SQS) Close() {
	close(q.ch)
}

type sqsConsumerMessage struct {
	data      []byte
	consumer  *sqs.Consumer
	id        *string
	logger    *zap.Logger
	expiredAt time.Time
	wg        *sync.WaitGroup
	ctx       context.Context
}

func (m *sqsConsumerMessage) Data() []byte {
	return m.data
}

func (m *sqsConsumerMessage) Done() {
	if err := m.consumer.DeleteMessage(m.id); err != nil {
		m.logger.Error("Error deleting message from SQS", zap.Error(err))
	}
	m.wg.Done()
}

func (m *sqsConsumerMessage) Failed() {
	m.wg.Done()
}

func (m *sqsConsumerMessage) IsExpired() bool {
	return m.expiredAt.Before(time.Now())
}
