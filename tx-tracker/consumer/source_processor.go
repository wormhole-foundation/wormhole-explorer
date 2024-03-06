package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var ErrAlreadyProcessed = errors.New("VAA was already processed")

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessSourceTxParams struct {
	TrackID   string
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
	Metrics   metrics.Metrics
}

func ProcessSourceTx(
	ctx context.Context,
	logger *zap.Logger,
	rpcPool map[vaa.ChainID]*pool.Pool,
	repository *Repository,
	params *ProcessSourceTxParams,
	p2pNetwork string,
) (*chains.TxDetail, error) {

	if !params.Overwrite {
		// If the message has already been processed, skip it.
		//
		// Sometimes the SQS visibility timeout expires and the message is put back into the queue,
		// even if the RPC nodes have been hit and data has been written to MongoDB.
		// In those cases, when we fetch the message for the second time,
		// we don't want to hit the RPC nodes again for performance reasons.
		processed, err := repository.AlreadyProcessed(ctx, params.VaaId)
		if err != nil {
			return nil, err
		} else if processed {
			return nil, ErrAlreadyProcessed
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

	if params.TxHash == "" {
		// add metrics for vaa without txHash
		params.Metrics.IncVaaWithoutTxHash(uint16(params.ChainId))
		v, err := repository.GetVaaIdTxHash(ctx, params.VaaId)
		if err != nil {
			logger.Error("failed to find vaaIdTxHash",
				zap.String("trackId", params.TrackID),
				zap.String("vaaId", params.VaaId),
				zap.Any("vaaTimestamp", params.Timestamp),
				zap.Error(err),
			)
		} else {
			// add metrics for vaa with txHash fixed
			params.Metrics.IncVaaWithTxHashFixed(uint16(params.ChainId))
			params.TxHash = v.TxHash
			logger.Warn("fix txHash for vaa",
				zap.String("trackId", params.TrackID),
				zap.String("vaaId", params.VaaId),
				zap.Any("vaaTimestamp", params.Timestamp),
				zap.String("txHash", v.TxHash),
			)
		}
	}

	if params.TxHash == "" {
		logger.Warn("txHash is empty",
			zap.String("trackId", params.TrackID),
			zap.String("vaaId", params.VaaId),
		)
		return nil, errors.New("txHash is empty")
	}

	// Get transaction details from the emitter blockchain
	txDetail, err = chains.FetchTx(ctx, rpcPool, params.ChainId, params.TxHash, params.Timestamp, p2pNetwork, params.Metrics, logger)
	if err != nil {
		return nil, err
	}

	// Store source transaction details in the database
	p := UpsertOriginTxParams{
		VaaId:     params.VaaId,
		TrackID:   params.TrackID,
		ChainId:   params.ChainId,
		Timestamp: params.Timestamp,
		TxDetail:  txDetail,
		TxStatus:  domain.SourceTxStatusConfirmed,
	}

	err = repository.UpsertOriginTx(ctx, &p)
	if err != nil {
		return nil, err
	}
	return txDetail, nil
}
