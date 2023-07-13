package consumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var ErrAlreadyProcessed = errors.New("VAA was already processed")

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessSourceTxParams struct {
	ChainId  sdk.ChainID
	VaaId    string
	Emitter  string
	Sequence string
	TxHash   string
	// Overwrite indicates whether to reprocess a VAA that has already been processed.
	//
	// In the context of backfilling, sometimes you want to overwrite old data (e.g.: because
	// the schema changed).
	// In the context of the service, you usually don't want to overwrite existing data
	// to avoid processing the same VAA twice, which would result in performance degradation.
	Overwrite bool
}

func ProcessSourceTx(
	ctx context.Context,
	logger *zap.Logger,
	rpcServiceProviderSettings *config.RpcProviderSettings,
	repository *Repository,
	params *ProcessSourceTxParams,
) error {

	if !params.Overwrite {
		// If the message has already been processed, skip it.
		//
		// Sometimes the SQS visibility timeout expires and the message is put back into the queue,
		// even if the RPC nodes have been hit and data has been written to MongoDB.
		// In those cases, when we fetch the message for the second time,
		// we don't want to hit the RPC nodes again for performance reasons.
		processed, err := repository.AlreadyProcessed(ctx, params.VaaId)
		if err != nil {
			return err
		} else if err == nil && processed {
			return ErrAlreadyProcessed
		}
	}

	// Get transaction details from the emitter blockchain
	txDetail, err := chains.FetchTx(ctx, rpcServiceProviderSettings, params.ChainId, params.TxHash)
	if err != nil {
		return fmt.Errorf("failed to process transaction: %w", err)
	}

	// Store source transaction details in the database
	p := UpsertDocumentParams{
		VaaId:    params.VaaId,
		ChainId:  params.ChainId,
		TxDetail: txDetail,
		TxStatus: domain.SourceTxStatusConfirmed,
	}
	return repository.UpsertDocument(ctx, &p)
}
