package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	sqs_client "github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	consumer  *sqs_client.Consumer
	ch        chan ConsumerMessage
	converter ConverterFunc
	chSize    int
	wg        sync.WaitGroup
	logger    *zap.Logger
}

// ConverterFunc converts a message from a sqs message.
type ConverterFunc func(string) (*Event, error)

// NewEventSqs creates a VAA queue in SQS instances.
func NewEventSqs(consumer *sqs_client.Consumer, converter ConverterFunc, logger *zap.Logger, opts ...SQSOption) *SQS {
	s := &SQS{
		consumer:  consumer,
		converter: converter,
		chSize:    10,
		logger:    logger.With(zap.String("queueUrl", consumer.GetQueueUrl())),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.ch = make(chan ConsumerMessage, s.chSize)
	return s
}

// WithChannelSize allows to specify an channel size when setting a value.
func WithChannelSize(size int) SQSOption {
	return func(d *SQS) {
		d.chSize = size
	}
}

// Consume returns the channel with the received messages from SQS queue.
func (q *SQS) Consume(ctx context.Context) <-chan ConsumerMessage {
	go func() {
		for {
			messages, err := q.consumer.GetMessages(ctx)
			if err != nil {
				q.logger.Error("Error getting messages from SQS", zap.Error(err))
				continue
			}
			q.logger.Debug("Received messages from SQS", zap.Int("count", len(messages)))
			expiredAt := time.Now().Add(q.consumer.GetVisibilityTimeout())
			for _, msg := range messages {
				// unmarshal body to sqsEvent
				var sqsEvent sqsEvent
				err := json.Unmarshal([]byte(*msg.Body), &sqsEvent)
				if err != nil {
					q.logger.Error("Error decoding message from SQS", zap.Error(err), zap.String("body", *msg.Body))
					if err = q.consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
						q.logger.Error("Error deleting message from SQS", zap.Error(err))
					}
					continue
				}

				// converts message to event
				event, err := q.converter(sqsEvent.Message)
				if err != nil {
					q.logger.Error("Error converting event message", zap.Error(err), zap.String("body", *msg.Body))
					if err = q.consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
						q.logger.Error("Error deleting message from SQS", zap.Error(err))
					}
					continue
				}

				if event == nil {
					q.logger.Warn("Can not handle message", zap.String("body", *msg.Body))
					if err = q.consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
						q.logger.Error("Error deleting message from SQS", zap.Error(err))
					}
					continue
				}

				retry, _ := strconv.Atoi(msg.Attributes["ApproximateReceiveCount"])
				q.wg.Add(1)
				q.ch <- &sqsConsumerMessage{
					id:            msg.ReceiptHandle,
					data:          event,
					wg:            &q.wg,
					logger:        q.logger,
					consumer:      q.consumer,
					retry:         uint8(retry),
					expiredAt:     expiredAt,
					sentTimestamp: sqs_client.GetSentTimestamp(msg),
					ctx:           ctx,
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
	data          *Event
	consumer      *sqs_client.Consumer
	wg            *sync.WaitGroup
	id            *string
	logger        *zap.Logger
	retry         uint8
	expiredAt     time.Time
	sentTimestamp *time.Time
	ctx           context.Context
}

func (m *sqsConsumerMessage) Data() *Event {
	return m.data
}

func (m *sqsConsumerMessage) Done() {
	if err := m.consumer.DeleteMessage(m.ctx, m.id); err != nil {
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

func (m *sqsConsumerMessage) Retry() uint8 {
	return m.retry
}

func (m *sqsConsumerMessage) SentTimestamp() *time.Time {
	return m.sentTimestamp
}
