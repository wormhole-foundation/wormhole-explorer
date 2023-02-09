package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume    queue.VAAConsumeFunc
	repository *parser.Repository
	parser     parser.ParserVAAAPIClient
	logger     *zap.Logger
}

// New creates a new vaa consumer.
func New(consume queue.VAAConsumeFunc, repository *parser.Repository, parser parser.ParserVAAAPIClient, logger *zap.Logger) *Consumer {
	return &Consumer{consume: consume, repository: repository, parser: parser, logger: logger}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			event := msg.Data()

			// check id message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Message with vaa expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			// unmarshal vaa.
			vaa, err := vaa.Unmarshal(event.Vaa)
			if err != nil {
				c.logger.Error("Invalid vaa", zap.String("id", event.ID), zap.Error(err))
				msg.Failed()
				continue
			}

			// call vaa-payload-parser api to parse a VAA.
			vaaParseResponse, err := c.parser.Parse(event.ChainID, event.EmitterAddress, event.Sequence, vaa.Payload)
			if err != nil {
				if errors.Is(err, parser.ErrInternalError) {
					c.logger.Info("error parsing VAA, will retry later", zap.Uint16("chainID", event.ChainID),
						zap.String("address", event.EmitterAddress), zap.String("sequence", event.Sequence), zap.Error(err))
					msg.Failed()
					continue
				}

				c.logger.Info("VAA cannot be parsed", zap.Uint16("chainID", event.ChainID),
					zap.String("address", event.EmitterAddress), zap.String("sequence", event.Sequence), zap.Error(err))
				msg.Done()
				continue
			}

			// create ParsedVaaUpdate to upsert.
			now := time.Now()
			vaaParsed := parser.ParsedVaaUpdate{
				ID:           event.ID,
				EmitterChain: event.ChainID,
				EmitterAddr:  event.EmitterAddress,
				Sequence:     event.Sequence,
				AppID:        vaaParseResponse.AppID,
				Result:       vaaParseResponse.Result,
				UpdatedAt:    &now,
			}

			err = c.repository.UpsertParsedVaa(ctx, vaaParsed)
			if err != nil {
				c.logger.Error("Error inserting vaa in repository",
					zap.String("id", event.ID),
					zap.Error(err))
				msg.Failed()
				continue
			}
			msg.Done()
			c.logger.Info("Vaa save in repository", zap.String("id", event.ID))
		}
	}()
}
