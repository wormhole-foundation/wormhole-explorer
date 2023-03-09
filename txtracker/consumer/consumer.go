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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type TxStatus uint

const (
	TxStatusChainNotSupported TxStatus = 0
	TxStatusFailedToProcess   TxStatus = 1
	TxStatusConfirmed         TxStatus = 2
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc        queue.VAAConsumeFunc
	cfg                *config.Settings
	logger             *zap.Logger
	globalTransactions *mongo.Collection
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	cfg *config.Settings,
	logger *zap.Logger,
	db *mongo.Database,
) *Consumer {

	c := Consumer{
		consumeFunc:        consumeFunc,
		cfg:                cfg,
		logger:             logger,
		globalTransactions: db.Collection("globalTransactions"),
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
			txStatus := TxStatusConfirmed
			txDetail, err := chains.FetchTx(ctx, c.cfg, event.ChainID, event.TxHash)
			if err == chains.ErrChainNotSupported {
				c.logger.Debug("Failed to fetch source transaction details - chain not supported",
					zap.String("vaaId", event.ID),
				)
				txStatus = TxStatusChainNotSupported
			} else if err != nil {
				c.logger.Error("Failed to fetch source transaction details",
					zap.String("vaaId", event.ID),
					zap.Error(err),
				)
				txStatus = TxStatusFailedToProcess
			}

			// store source transaction details in the database
			err = updateSourceTxData(ctx, c.globalTransactions, event, txDetail, txStatus)
			if err != nil {
				c.logger.Error("Failed to upsert source transaction details",
					zap.String("vaaId", event.ID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Successfuly updated source transaction details in the database",
					zap.String("id", event.ID),
					zap.Any("details", txDetail),
				)
			}

			msg.Done()
		}
	}()
}

func updateSourceTxData(
	ctx context.Context,
	vaas *mongo.Collection,
	event *queue.VaaEvent,
	txDetail *chains.TxDetail,
	txStatus TxStatus,
) error {

	fields := bson.D{
		{Key: "chainId", Value: event.ChainID},
		{Key: "txHash", Value: event.TxHash},
		{Key: "status", Value: txStatus},
	}

	if txDetail != nil {
		fields = append(fields, primitive.E{Key: "nativeTxHash", Value: txDetail.NativeTxHash})
		fields = append(fields, primitive.E{Key: "timestamp", Value: txDetail.Timestamp})
		fields = append(fields, primitive.E{Key: "signer", Value: txDetail.Signer})
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "originTx",
					Value: fields,
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
