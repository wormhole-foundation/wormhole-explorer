package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

	"go.uber.org/zap"
)

// VAAQueueConsumer represents a VAA queue consumer.
type VAAQueueConsumer struct {
	consume    VAAQueueConsumeFunc
	repository storage.Storager
	notifyFunc VAANotifyFunc
	metrics    metrics.Metrics
	logger     *zap.Logger
}

// NewVAAQueueConsumer creates a new VAA queue consumer instances.
func NewVAAQueueConsumer(
	consume VAAQueueConsumeFunc,
	repository storage.Storager,
	notifyFunc VAANotifyFunc,
	metrics metrics.Metrics,
	logger *zap.Logger) *VAAQueueConsumer {
	return &VAAQueueConsumer{
		consume:    consume,
		repository: repository,
		notifyFunc: notifyFunc,
		metrics:    metrics,
		logger:     logger,
	}
}

// Start consumes messages from VAA queue and store those messages in a repository.
func (c *VAAQueueConsumer) Start(ctx context.Context, runMode string) {
	if runMode == config.RunModeLegacy {
		c.legacy(ctx)
	} else {
		c.start(ctx)
	}
}

// Start consumes messages from VAA queue and store those messages in a mongo repository.
// TODO: remove after migration.
func (c *VAAQueueConsumer) legacy(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			v, err := sdk.Unmarshal(msg.Data())
			if err != nil {
				c.logger.Error("Error unmarshalling vaa", zap.Error(err))
				msg.Failed()
				continue
			}

			if msg.IsExpired() {
				c.logger.Warn("Message with vaa expired", zap.String("id", v.MessageID()))
				msg.Failed()
				continue
			}

			c.metrics.IncVaaConsumedFromQueue(v.EmitterChain)

			c.metrics.IncConsistencyLevelByChainID(v.EmitterChain, v.ConsistencyLevel)

			if v.EmitterChain != sdk.ChainIDPythNet && domain.ConsistencyLevelIsImmediately(v) {
				dbVaa, err := c.repository.FindVaaByID(ctx, v.MessageID())
				if err != nil {
					c.logger.Error("Error finding vaa in repository",
						zap.String("id", v.MessageID()),
						zap.Error(err))
					msg.Failed()
					continue
				}
				if dbVaa == nil {
					err = c.repository.UpsertVAA(ctx, v, msg.Data())
					if err != nil {
						c.logger.Error("Error inserting vaa in repository",
							zap.String("id", v.MessageID()),
							zap.Error(err))
						msg.Failed()
						continue
					}
				} else {
					existingVaa, err := sdk.Unmarshal(dbVaa.Vaa)
					if err != nil {
						c.logger.Error("Error unmarshalling found vaa", zap.Error(err), zap.String("id", v.MessageID()))
						msg.Failed()
						continue
					}
					currentHash := v.SigningDigest()
					savedHash := existingVaa.SigningDigest()
					// if the hash is the same, we can skip the vaa
					if currentHash.Hex() == savedHash.Hex() {
						msg.Done(ctx)
						continue
					}
					//put as dirty the vaa and save it in duplicatedVaas
					err = c.repository.UpsertDuplicateVaa(ctx, v, msg.Data())
					if err != nil {
						c.logger.Error("Error inserting duplicate vaa in repository",
							zap.String("id", v.MessageID()),
							zap.Error(err))
						msg.Failed()
						continue
					}
					c.metrics.IncDuplicateVaaByChainID(v.EmitterChain)
				}
			} else {
				err = c.repository.UpsertVAA(ctx, v, msg.Data())
				if err != nil {
					c.logger.Error("Error inserting vaa in repository",
						zap.String("id", v.MessageID()),
						zap.Error(err))
					msg.Failed()
					continue
				}
			}

			err = c.notifyFunc(ctx, v, msg.Data())
			if err != nil {
				c.metrics.IncMaxSequenceCacheError(v.EmitterChain)
				c.logger.Error("Error notifying vaa",
					zap.String("id", v.MessageID()),
					zap.Error(err))
				msg.Failed()
				continue
			}
			c.metrics.VaaProcessingDuration(v.EmitterChain, msg.SentTimestamp())
			msg.Done(ctx)
			c.logger.Info("Vaa saved in repository", zap.String("id", v.MessageID()))
		}
	}()
}

// Start consumes messages from VAA queue and store those messages in a repository.
func (c *VAAQueueConsumer) start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			v, err := sdk.Unmarshal(msg.Data())
			if err != nil {
				c.logger.Error("Error unmarshalling vaa", zap.Error(err))
				msg.Failed()
				continue
			}

			if msg.IsExpired() {
				c.logger.Warn("Message with vaa expired", zap.String("id", v.MessageID()))
				msg.Failed()
				continue
			}

			c.metrics.IncVaaConsumedFromQueue(v.EmitterChain)

			c.metrics.IncConsistencyLevelByChainID(v.EmitterChain, v.ConsistencyLevel)

			// upsert vaa in repository and dispatch events.
			err = c.repository.UpsertVAA(ctx, v, msg.Data())
			if err != nil {
				c.logger.Error("Error inserting vaa in repository",
					zap.String("id", v.MessageID()),
					zap.Error(err))
				msg.Failed()
				continue
			}

			// notify max sequence cache
			err = c.notifyFunc(ctx, v, msg.Data())
			if err != nil {
				c.metrics.IncMaxSequenceCacheError(v.EmitterChain)
				c.logger.Error("Error notifying vaa",
					zap.String("id", v.MessageID()),
					zap.Error(err))
				msg.Failed()
				continue
			}
			c.metrics.VaaProcessingDuration(v.EmitterChain, msg.SentTimestamp())
			msg.Done(ctx)
			c.logger.Info("Vaa saved in repository", zap.String("id", v.MessageID()))
		}
	}()
}
