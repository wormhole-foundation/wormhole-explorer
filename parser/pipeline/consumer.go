package pipeline

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"go.uber.org/zap"
)

type Consumer struct {
	consume    queue.VAAConsumeFunc
	repository *parser.Repository
	parser     parser.ParserVAAAPIClient
	logger     *zap.Logger
}

// NewConsumer creates a new vaa consumer.
func NewConsumer(consume queue.VAAConsumeFunc, repository *parser.Repository, parser parser.ParserVAAAPIClient, logger *zap.Logger) *Consumer {
	return &Consumer{consume: consume, repository: repository, parser: parser, logger: logger}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.consume(ctx):
				event := msg.Data

				// check id message is expired.
				if msg.IsExpired() {
					c.logger.Warn("Message with vaa expired", zap.String("id", event.ID()))
					continue
				}

				// call vaa-payload-parser api to parse a VAA.
				sequence := strconv.FormatUint(event.Sequence, 10)
				vaaParseResponse, err := c.parser.Parse(event.ChainID, event.EmitterAddress, sequence, event.Vaa)
				if err != nil {
					if errors.Is(err, parser.ErrInternalError) {
						c.logger.Info("error parsing VAA, will retry later", zap.Uint16("chainID", event.ChainID),
							zap.String("address", event.EmitterAddress), zap.Uint64("sequence", event.Sequence), zap.Error(err))
						continue
					}

					c.logger.Info("VAA cannot be parsed", zap.Uint16("chainID", event.ChainID),
						zap.String("address", event.EmitterAddress), zap.Uint64("sequence", event.Sequence), zap.Error(err))
					msg.Ack()
					continue
				}

				// create ParsedVaaUpdate to upsert.
				now := time.Now()
				vaaParsed := parser.ParsedVaaUpdate{
					ID:           event.ID(),
					EmitterChain: event.ChainID,
					EmitterAddr:  event.EmitterAddress,
					Sequence:     strconv.FormatUint(event.Sequence, 10),
					AppID:        vaaParseResponse.AppID,
					Result:       vaaParseResponse.Result,
					UpdatedAt:    &now,
				}

				err = c.repository.UpsertParsedVaa(ctx, vaaParsed)
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
