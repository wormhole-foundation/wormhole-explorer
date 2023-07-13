package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var ErrAlreadyProcessed = errors.New("VAA was already processed")

const (
	minRetries    = 3
	retryDelay    = 1 * time.Minute
	retryDeadline = 10 * time.Minute
)

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessSourceTxParams struct {
	Timestamp *time.Time
	ChainId   sdk.ChainID
	VaaId     string
	Emitter   string
	Sequence  string
	TxHash    string
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

	// The loop below tries to fetch transaction details from an external API / RPC node.
	//
	// It keeps retrying until both of these conditions are met:
	// 1. A fixed amount of time has passed since the VAA was emitted (this is because
	//    some chains have awful finality times).
	// 2. A minimum number of attempts have been made.
	var txDetail *chains.TxDetail
	var err error
	for retries := 0; ; retries++ {

		// Get transaction details from the emitter blockchain
		txDetail, err = chains.FetchTx(ctx, rpcServiceProviderSettings, params.ChainId, params.TxHash)
		if err == nil {
			break
		}

		// Keep retrying?
		if params.Timestamp == nil && retries > minRetries {
			return fmt.Errorf("failed to process transaction: %w", err)
		} else if time.Since(*params.Timestamp) > retryDeadline && retries >= minRetries {
			return fmt.Errorf("failed to process transaction: %w", err)
		} else {
			logger.Warn("failed to process transaction",
				zap.Any("vaaTimestamp", params.Timestamp),
				zap.Int("retries", retries),
				zap.Error(err),
			)
			if params.Timestamp != nil && time.Since(*params.Timestamp) < retryDeadline {
				time.Sleep(retryDelay)
			}
		}
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
