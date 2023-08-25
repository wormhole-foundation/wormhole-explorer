package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	sqs_client "github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// SQSOption represents a VAA queue in SQS option function.
type SQSOption func(*SQS)

// SQS represents a VAA queue in SQS.
type SQS struct {
	consumer *sqs_client.Consumer
	ch       chan ConsumerMessage
	chSize   int
	wg       sync.WaitGroup
	metrics  metrics.Metrics
	logger   *zap.Logger
}

// FilterConsumeFunc filter vaaa func definition.
type FilterConsumeFunc func(vaaEvent *Event) bool

// NewVaaSqs creates a VAA queue in SQS instances.
func NewVaaSqs(consumer *sqs_client.Consumer, metrics metrics.Metrics, logger *zap.Logger, opts ...SQSOption) *SQS {
	s := &SQS{
		consumer: consumer,
		chSize:   10,
		metrics:  metrics,
		logger:   logger}
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
			expiredAt := time.Now().Add(q.consumer.GetVisibilityTimeout())
			for _, msg := range messages {
				// unmarshal body to sqsEvent
				var sqsEvent sqsEvent
				err := json.Unmarshal([]byte(*msg.Body), &sqsEvent)
				if err != nil {
					q.logger.Error("Error decoding message from SQS", zap.Error(err))
					continue
				}

				// unmarshal message to NotificationEvent
				var notification domain.NotificationEvent
				err = json.Unmarshal([]byte(sqsEvent.Message), &notification)
				if err != nil {
					q.logger.Error("Error decoding vaaEvent message from SQSEvent", zap.Error(err))
					continue
				}

				event := q.createEvent(&notification)
				if event == nil {
					continue
				}

				q.metrics.IncVaaConsumedQueue(uint16(event.ChainID))

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

func (q *SQS) createEvent(notification *domain.NotificationEvent) *Event {
	if notification.Type != domain.SignedVaaType && notification.Type != domain.PublishedLogMessageType {
		q.logger.Debug("Skip event type", zap.String("trackId", notification.TrackID), zap.String("type", notification.Type))
		return nil
	}

	switch notification.Type {
	case domain.SignedVaaType:
		signedVaa, err := domain.GetEventPayload[domain.SignedVaa](notification)
		if err != nil {
			q.logger.Error("Error decoding signedVAA from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
			return nil
		}
		return &Event{
			ID:             signedVaa.ID,
			ChainID:        sdk.ChainID(signedVaa.EmitterChain),
			EmitterAddress: signedVaa.EmitterAddr,
			Sequence:       fmt.Sprintf("%d", signedVaa.Sequence),
			Vaa:            signedVaa.Vaa,
			Timestamp:      &signedVaa.Timestamp,
			TxHash:         signedVaa.TxHash,
		}
	case domain.PublishedLogMessageType:
		plm, err := domain.GetEventPayload[domain.PublishedLogMessage](notification)
		if err != nil {
			q.logger.Error("Error decoding publishedLogMessage from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
			return nil
		}
		return &Event{
			ID:             plm.ID,
			ChainID:        sdk.ChainID(plm.EmitterChain),
			EmitterAddress: plm.EmitterAddr,
			Sequence:       strconv.FormatUint(plm.Sequence, 10),
			Vaa:            plm.Vaa,
			Timestamp:      &plm.Timestamp,
			TxHash:         plm.TxHash,
		}
	}
	return nil
}

type sqsConsumerMessage struct {
	data      *Event
	consumer  *sqs_client.Consumer
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
		m.logger.Error("Error deleting message from SQS",
			zap.String("vaaId", m.data.ID),
			zap.Bool("isExpired", m.IsExpired()),
			zap.Time("expiredAt", m.expiredAt),
			zap.Error(err),
		)
	}
	m.wg.Done()
}

func (m *sqsConsumerMessage) Failed() {
	m.wg.Done()
}

func (m *sqsConsumerMessage) IsExpired() bool {
	return m.expiredAt.Before(time.Now())
}
