package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

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

		// If the context was cancelled, do not attempt to save the result on the database
		case errors.Is(err, context.Canceled):
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
