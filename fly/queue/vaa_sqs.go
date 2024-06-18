package queue

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	common_sqs "github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// VAASqsOption represents a VAA queue in SQS option function.
type VAASqsOption func(*VAASqs)

// VAASqs represents a VAA queue in VAASqs.
type VAASqs struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan Message[[]byte]
	chSize   int
	wg       sync.WaitGroup
	logger   *zap.Logger
}

// NewVaaSqs creates a VAA queue in SQS instances.
func NewVaaSqs(producer *sqs.Producer, consumer *sqs.Consumer, logger *zap.Logger, opts ...VAASqsOption) *VAASqs {
	s := &VAASqs{
		producer: producer,
		consumer: consumer,
		chSize:   10,
		logger:   logger.With(zap.String("queueUrl", consumer.GetQueueUrl()))}
	for _, opt := range opts {
		opt(s)
	}
	s.ch = make(chan Message[[]byte], s.chSize)
	return s
}

// WithChannelSize allows to specify an channel size when setting a value.
func WithChannelSize(size int) VAASqsOption {
	return func(d *VAASqs) {
		d.chSize = size
	}
}

// Publish sends the message to a SQS queue.
func (q *VAASqs) Publish(ctx context.Context, v *sdk.VAA, data []byte) error {
	body := base64.StdEncoding.EncodeToString(data)
	deduplicationID := createVaaDeduplicationID(v)
	return q.producer.SendMessage(ctx, deduplicationID, deduplicationID, body)
}

// Consume returns the channel with the received messages from SQS queue.
func (q *VAASqs) Consume(ctx context.Context) <-chan Message[[]byte] {
	go func() {
		for {
			messages, err := q.consumer.GetMessages(ctx)
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
				q.ch <- &sqsConsumerMessage[[]byte]{
					id:            msg.ReceiptHandle,
					data:          body,
					wg:            &q.wg,
					logger:        q.logger,
					consumer:      q.consumer,
					expiredAt:     expiredAt,
					ctx:           ctx,
					sentTimestamp: common_sqs.GetSentTimestamp(msg),
				}
			}
			q.wg.Wait()
		}
	}()
	return q.ch
}

// Close closes all consumer resources.
func (q *VAASqs) Close() {
	close(q.ch)
}

func createVaaDeduplicationID(v *sdk.VAA) string {
	deduplicationID := domain.CreateUniqueVaaID(v)
	if len(deduplicationID) > 127 {
		return deduplicationID[:127]
	}
	return deduplicationID
}
