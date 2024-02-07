package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"

	"go.uber.org/zap"
)

// ObservationQueueConsumer represents a observation queue consumer.
type ObservationQueueConsumer struct {
	consume    ObservationQueueConsumeFunc
	repository *storage.Repository
	metrics    metrics.Metrics
	logger     *zap.Logger
}

// ObservationQueueConsumer creates a new observation queue consumer instances.
func NewObservationQueueConsumer(
	consume ObservationQueueConsumeFunc,
	repository *storage.Repository,
	metrics metrics.Metrics,
	logger *zap.Logger) *ObservationQueueConsumer {
	return &ObservationQueueConsumer{
		consume:    consume,
		repository: repository,
		metrics:    metrics,
		logger:     logger,
	}
}

// Start consumes messages from observation queue and store those messages in a repository.
func (c *ObservationQueueConsumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			obs := msg.Data()
			log := c.logger.With(zap.String("id", obs.MessageId))
			log.Info("Observation message received")

			if msg.IsExpired() {
				log.Warn("Message with observation expired")
				msg.Failed()
				continue
			}
			err := c.repository.UpsertObservation(ctx, obs)
			if err != nil {
				log.Error("Error inserting observation in repository", zap.Error(err))
				msg.Failed()
				continue
			}
			msg.Done(ctx)
			c.logger.Info("Observation saved in repository")
		}
	}()
}
