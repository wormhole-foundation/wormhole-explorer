package pipeline

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"go.uber.org/zap"
)

type Consumer struct {
	consume    queue.VAAConsumeFunc
	repository *parser.Repository
	parser     *parser.NodeJS
	logger     *zap.Logger
}

func NewConsumer(consume queue.VAAConsumeFunc, repository *parser.Repository, parser *parser.NodeJS, logger *zap.Logger) *Consumer {
	return &Consumer{consume: consume, repository: repository, parser: parser, logger: logger}
}

// Start consumes messages from VAA queue and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.consume(ctx):
				event := msg.Data
				vpf, err := c.repository.GetVaaParserFunction(ctx, event.ChainID, event.EmitterAddress)
				if err != nil {
					c.logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}

				result, err := c.parser.Parse(vpf.ParserFunction, event.Vaa)
				if err != nil {
					c.logger.Error("Error parsing vaa", zap.Error(err))
					continue
				}

				if msg.IsExpired() {
					c.logger.Warn("Message with vaa expired", zap.String("id", event.ID()))
					continue
				}

				err = c.repository.UpsertParsedVaa(ctx, event, result)
				if err != nil {
					c.logger.Error("Error inserting vaa in repository",
						zap.String("id", event.ID()),
						zap.Error(err))
					continue
				}
				msg.Ack()
				c.logger.Info("Vaa save in repository", zap.String("id", event.ID()))
			}
		}
	}()
}
