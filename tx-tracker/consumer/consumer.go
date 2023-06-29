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
	workerPool                 *WorkerPool
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	rpcServiceProviderSettings *config.RpcProviderSettings,
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
) *Consumer {

	workerPool := NewWorkerPool(ctx, logger, rpcServiceProviderSettings, repository)

	c := Consumer{
		consumeFunc:                consumeFunc,
		rpcServiceProviderSettings: rpcServiceProviderSettings,
		logger:                     logger,
		repository:                 repository,
		workerPool:                 workerPool,
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go c.producerLoop(ctx)
}

func (c *Consumer) producerLoop(ctx context.Context) {

	ch := c.consumeFunc(ctx)

	for msg := range ch {

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

		// Send the VAA to the worker pool.
		p := ProcessSourceTxParams{
			VaaId:    event.ID,
			ChainId:  event.ChainID,
			Emitter:  event.EmitterAddress,
			Sequence: event.Sequence,
			TxHash:   event.TxHash,
		}
		err := c.workerPool.Push(ctx, &p)
		if err != nil {
			c.logger.Warn("failed to push message into worker pool",
				zap.String("vaaId", event.ID),
				zap.Error(err),
			)
			msg.Failed()
		}

		msg.Done()
	}
}

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessSourceTxParams struct {
	ChainId  sdk.ChainID
	VaaId    string
	Emitter  string
	Sequence string
	TxHash   string
}

func ProcessSourceTx(
	ctx context.Context,
	logger *zap.Logger,
	rpcServiceProviderSettings *config.RpcProviderSettings,
	repository *Repository,
	params *ProcessSourceTxParams,
) error {

	// Get transaction details from the emitter blockchain
	//
	// If the transaction is not found, will retry a few times before giving up.
	var txStatus domain.SourceTxStatus
	var txDetail *chains.TxDetail
	var err error
	for attempts := numRetries; attempts > 0; attempts-- {

		txDetail, err = chains.FetchTx(ctx, rpcServiceProviderSettings, params.ChainId, params.TxHash)

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
			logger.Error("Failed to fetch source transaction details",
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
	return repository.UpsertDocument(ctx, &p)
}
