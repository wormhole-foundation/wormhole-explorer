package queue

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"go.uber.org/zap"
)

// ObservationSqs represents a observation queue in SQS.
type ObservationSqs struct {
	producer *sqs.Producer
	consumer *sqs.Consumer
	ch       chan Message[*gossipv1.SignedObservation]
	chSize   int
	wg       sync.WaitGroup
	logger   *zap.Logger
}

// NewObservationSqs creates a observation queue in SQS instances.
func NewObservationSqs(producer *sqs.Producer, consumer *sqs.Consumer, logger *zap.Logger, opts ...VAASqsOption) *ObservationSqs {
	s := &ObservationSqs{
		producer: producer,
		consumer: consumer,
		chSize:   10,
		logger:   logger.With(zap.String("queueUrl", consumer.GetQueueUrl()))}
	s.ch = make(chan Message[*gossipv1.SignedObservation], s.chSize)
	return s
}

// Publish sends the message to a SQS queue.
func (q *ObservationSqs) Publish(ctx context.Context, o *gossipv1.SignedObservation) error {
	dto := toObservation(o)
	body, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s/%s/%s", o.MessageId, hex.EncodeToString(o.Addr), hex.EncodeToString(o.Hash))
	deduplicationId := base64.StdEncoding.EncodeToString([]byte(id))[:127]
	return q.producer.SendMessage(ctx, deduplicationId, deduplicationId, string(body))
}

// Consume returns the channel with the received messages from SQS queue.
func (q *ObservationSqs) Consume(ctx context.Context) <-chan Message[*gossipv1.SignedObservation] {
	go func() {
		for {
			messages, err := q.consumer.GetMessages(ctx)
			if err != nil {
				q.logger.Error("Error getting messages from SQS", zap.Error(err))
				continue
			}
			q.logger.Info("Received messages from SQS", zap.Int("count", len(messages)))
			expiredAt := time.Now().Add(q.consumer.GetVisibilityTimeout())
			for _, msg := range messages {
				var obs Observation
				err := json.Unmarshal([]byte(*msg.Body), &obs)
				if err != nil {
					q.logger.Error("Error decoding message from SQS", zap.Error(err))
					continue
				}
				q.logger.Info("Observation message received", zap.String("id", obs.MessageID))

				//TODO check if callback is better than channel
				q.wg.Add(1)
				q.ch <- &sqsConsumerMessage[*gossipv1.SignedObservation]{
					id:        msg.ReceiptHandle,
					data:      fromObservation(&obs),
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
func (q *ObservationSqs) Close() {
	close(q.ch)
}
