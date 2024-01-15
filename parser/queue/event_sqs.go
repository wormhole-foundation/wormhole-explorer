package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"go.uber.org/zap"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	consumer      *sqs.Consumer
	ch            chan ConsumerMessage
	chSize        int
	wg            sync.WaitGroup
	filterConsume FilterConsumeFunc
	converter     ConverterFunc
	metrics       metrics.Metrics
	logger        *zap.Logger
}

// FilterConsumeFunc filter vaaa func definition.
type FilterConsumeFunc func(*Event) bool

// ConverterFunc converts a message from a sqs message.
type ConverterFunc func(string) (*Event, error)

// NewEventSQS creates a VAA queue in SQS instances.
func NewEventSQS(consumer *sqs.Consumer, converter ConverterFunc, filterConsume FilterConsumeFunc, metrics metrics.Metrics, logger *zap.Logger, opts ...SQSOption) *SQS {
	s := &SQS{
		consumer:      consumer,
		chSize:        10,
		converter:     converter,
		filterConsume: filterConsume,
		metrics:       metrics,
		logger:        logger.With(zap.String("queueUrl", consumer.GetQueueUrl())),
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
					q.logger.Error("Error decoding message from SQS", zap.Error(err))
					if err = q.consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
						q.logger.Error("Error deleting message from SQS", zap.Error(err))
					}
					continue
				}

				// unmarshal message to event
				event, err := q.converter(sqsEvent.Message)
				if err != nil {
					q.logger.Error("Error converting event message", zap.Error(err))
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

				q.metrics.IncVaaConsumedQueue(event.ChainID)

				// filter vaaEvent by p2p net.
				if q.filterConsume(event) {
					if err := q.consumer.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
						q.logger.Error("Error deleting message from SQS", zap.Error(err))
					}
					continue
				}
				q.metrics.IncVaaUnfiltered(event.ChainID)

				q.wg.Add(1)
				q.ch <- &sqsConsumerMessage{
					id:        msg.ReceiptHandle,
					data:      event,
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
	data      *Event
	consumer  *sqs.Consumer
	wg        *sync.WaitGroup
	id        *string
	logger    *zap.Logger
	expiredAt time.Time
	ctx       context.Context
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
