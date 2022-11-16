package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VAAQueueConsumeFunc func(context.Context) <-chan *queue.Message

type VAAQueueConsumer struct {
	consume    VAAQueueConsumeFunc
	repository *storage.Repository
	logger     *zap.Logger
}

func NewVAAQueueConsumer(
	consume VAAQueueConsumeFunc,
	repository *storage.Repository,
	logger *zap.Logger) *VAAQueueConsumer {
	return &VAAQueueConsumer{
		consume:    consume,
		repository: repository,
		logger:     logger,
	}
}

func (c *VAAQueueConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.consume(ctx):
				v, err := vaa.Unmarshal(msg.Data)
				if err != nil {
					c.logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}
				err = c.repository.UpsertVaa(ctx, v, msg.Data)
				if err != nil {
					c.logger.Error("Error inserting vaa in repository",
						zap.String("id", v.MessageID()),
						zap.Error(err))
					continue
				}
				msg.Ack()
				c.logger.Info("Vaa save in repository", zap.String("id", v.MessageID()))
			}
		}
	}()
}
