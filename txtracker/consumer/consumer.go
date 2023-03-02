package consumer

import (
	"context"
	"fmt"

	//"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc queue.VAAConsumeFunc
	cfg         *config.Settings
	logger      *zap.Logger
	vaas        *mongo.Collection
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	cfg *config.Settings,
	logger *zap.Logger,
	db *mongo.Database,
) *Consumer {

	c := Consumer{
		consumeFunc: consumeFunc,
		cfg:         cfg,
		logger:      logger,
		vaas:        db.Collection("vaas"),
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consumeFunc(ctx) {
			event := msg.Data()

			// check if message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Message with VAA expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			// do not process messages from PythNet
			if event.ChainID == sdk.ChainIDPythNet {
				msg.Done()
				continue
			}

			// get transaction details from the emitter blockchain
			txDetail, err := chains.FetchTx(ctx, c.cfg, event.ChainID, event.TxHash)
			if err == chains.ErrChainNotSupported {
				c.logger.Debug("Failed to fetch source transaction details - chain not supported",
					zap.String("vaaId", event.ID),
				)
				msg.Done()
				continue
			} else if err != nil {
				c.logger.Error("Failed to fetch source transaction details",
					zap.String("vaaId", event.ID),
					zap.Error(err),
				)
				msg.Done()
				continue
			}
			c.logger.Debug("Successfuly obtained source transaction details",
				zap.String("id", event.ID),
				zap.Any("details", txDetail),
			)

			// store source transaction details in the database
			err = updateSourceTxData(ctx, c.vaas, event, txDetail)
			if err != nil {
				c.logger.Error("Failed to upsert source transaction details",
					zap.String("vaaId", event.ID),
					zap.Error(err),
				)
				msg.Done()
				continue
			}
			c.logger.Debug("Successfuly updated source transaction details in the database",
				zap.String("id", event.ID),
				zap.Any("details", txDetail),
			)

			msg.Done()
		}
	}()
}

func updateSourceTxData(
	ctx context.Context,
	vaas *mongo.Collection,
	event *queue.VaaEvent,
	txDetail *chains.TxDetail,
) error {

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key: "metadata.sourceTx",
					Value: bson.D{
						{Key: "timestamp", Value: txDetail.Timestamp},
						{Key: "sender", Value: txDetail.Source},
						{Key: "receiver", Value: txDetail.Destination},
					},
				},
			},
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := vaas.UpdateByID(ctx, event.ID, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert source tx information: %w", err)
	}

	return nil
}
