package consumer

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	numRetries = 2
	retryDelay = 10 * time.Second
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc                queue.VAAConsumeFunc
	rpcServiceProviderSettings *config.RpcProviderSettings
	logger                     *zap.Logger
	repository                 *Repository
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	vaaPayloadParserSettings *config.VaaPayloadParserSettings,
	rpcServiceProviderSettings *config.RpcProviderSettings,
	logger *zap.Logger,
	repository *Repository,
) (*Consumer, error) {

	c := Consumer{
		consumeFunc:                consumeFunc,
		rpcServiceProviderSettings: rpcServiceProviderSettings,
		logger:                     logger,
		repository:                 repository,
	}

	return &c, nil
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consumeFunc(ctx) {
			event := msg.Data()

			// Check if message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Message with VAA expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			// Do not process messages from PythNet
			if event.ChainID == sdk.ChainIDPythNet {
				msg.Done()
				continue
			}

			// Fetch tx details from the corresponding RPC/API, then persist them on MongoDB.
			p := ProcessSourceTxParams{
				VaaId:    event.ID,
				ChainId:  event.ChainID,
				Emitter:  event.EmitterAddress,
				Sequence: event.Sequence,
				TxHash:   event.TxHash,
			}
			err := c.ProcessSourceTx(ctx, &p)
			if err == chains.ErrChainNotSupported {
				c.logger.Debug("Skipping VAA - chain not supported",
					zap.String("vaaId", event.ID),
				)
			} else if err != nil {
				c.logger.Error("Failed to upsert source transaction details",
					zap.String("vaaId", event.ID),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Updated source transaction details in the database",
					zap.String("id", event.ID),
				)
			}

			msg.Done()
		}
	}()
}

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessSourceTxParams struct {
	ChainId  sdk.ChainID
	VaaId    string
	Emitter  string
	Sequence string
	TxHash   string
}

func (c *Consumer) ProcessSourceTx(
	ctx context.Context,
	params *ProcessSourceTxParams,
) error {

	// Get transaction details from the emitter blockchain
	//
	// If the transaction is not found, will retry a few times before giving up.
	var txStatus domain.SourceTxStatus
	var txDetail *chains.TxDetail
	var err error
	for attempts := numRetries; attempts > 0; attempts-- {

		txDetail, err = chains.FetchTx(ctx, c.rpcServiceProviderSettings, params.ChainId, params.TxHash)

		switch {
		// If the transaction is not found, retry after a delay
		case err == chains.ErrTransactionNotFound:
			txStatus = domain.SourceTxStatusInternalError
			time.Sleep(retryDelay)
			continue

		// If the chain ID is not supported, we're done.
		case err == chains.ErrChainNotSupported:
			return err

		// If there is an internal error, give up
		case err != nil:
			c.logger.Error("Failed to fetch source transaction details",
				zap.String("vaaId", params.VaaId),
				zap.Error(err),
			)
			txStatus = domain.SourceTxStatusInternalError
			break

		// Success
		case err == nil:
			txStatus = domain.SourceTxStatusConfirmed
			break
		}
	}

	// Store source transaction details in the database
	p := UpsertDocumentParams{
		VaaId:    params.VaaId,
		ChainId:  params.ChainId,
		TxHash:   params.TxHash,
		TxDetail: txDetail,
		TxStatus: txStatus,
	}
	return c.repository.UpsertDocument(ctx, &p)
}
