package processor

import (
	"context"
	"fly/queue"
	"fly/storage"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VAAConsume func() <-chan *queue.Message

type VAAConsumer struct {
	consume    VAAConsume
	repository *storage.Repository
	logger     *zap.Logger
}

func NewVAAConsumer(
	consume VAAConsume,
	repository *storage.Repository,
	logger *zap.Logger) *VAAConsumer {
	return &VAAConsumer{
		consume:    consume,
		repository: repository,
		logger:     logger,
	}
}

func (c *VAAConsumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.consume():
				v, err := vaa.Unmarshal(msg.Data)
				if err != nil {
					c.logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}
				err = c.repository.UpsertVaa(v, msg.Data)
				if err != nil {
					c.logger.Error("Error inserting vaa in repository", zap.Error(err))
					continue
				}
				msg.Ack()
				c.logger.Info("Vaa save in repository", zap.String("id", v.MessageID()))
			}
		}
	}()
}
