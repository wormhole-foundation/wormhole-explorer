package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var (
	errTxFailedCannotBeUpdated = errors.New("tx with status failed can not be updated because exists a confirmed tx for the same vaa ID")
	errTxUnknowCannotBeUpdated = errors.New("tx with status unknown can not be updated because exists a tx (confirmed|failed) for the same vaa ID")
	errInvalidTxStatus         = errors.New("invalid tx status")
)

// ProcessSourceTxParams is a struct that contains the parameters for the ProcessSourceTx method.
type ProcessTargetTxParams struct {
	TrackID        string
	VaaId          string
	ChainId        sdk.ChainID
	Emitter        string
	TxHash         string
	BlockTimestamp *time.Time
	BlockHeight    string
	Method         string
	From           string
	To             string
	Status         string
}

func ProcessTargetTx(
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
	params *ProcessTargetTxParams,
) error {

	now := time.Now()
	update := &TargetTxUpdate{
		ID:      params.VaaId,
		TrackID: params.TrackID,
		Destination: &DestinationTx{
			ChainID:     params.ChainId,
			Status:      params.Status,
			TxHash:      params.TxHash,
			BlockNumber: params.BlockHeight,
			Timestamp:   params.BlockTimestamp,
			From:        params.From,
			To:          params.To,
			Method:      params.Method,
			UpdatedAt:   &now,
		},
	}

	// check if the transaction should be updated.
	shoudBeUpdated, err := checkTxShouldBeUpdated(ctx, update, repository)
	if !shoudBeUpdated {
		logger.Warn("Transaction should not be updated", zap.String("vaaId", params.VaaId), zap.Error(err))
		return nil
	}
	return repository.UpsertTargetTx(ctx, update)
}

func checkTxShouldBeUpdated(ctx context.Context, tx *TargetTxUpdate, repository *Repository) (bool, error) {
	switch tx.Destination.Status {
	case domain.DstTxStatusConfirmed:
		return true, nil
	case domain.DstTxStatusFailedToProcess:
		// check if the transaction exists from the same vaa ID.
		oldTx, err := repository.GetTargetTx(ctx, tx.ID)
		if err != nil {
			return true, nil
		}
		// if the transaction was already confirmed, then no update it.
		if oldTx != nil && oldTx.Destination.Status == domain.DstTxStatusConfirmed {
			return false, errTxFailedCannotBeUpdated
		}
		return true, nil
	case domain.DstTxStatusUnkonwn:
		// check if the transaction exists from the same vaa ID.
		oldTx, err := repository.GetTargetTx(ctx, tx.ID)
		if err != nil {
			return true, nil
		}
		// if the transaction was already confirmed or failed to process, then no update it.
		if oldTx.Destination.Status == domain.DstTxStatusConfirmed || oldTx.Destination.Status == domain.DstTxStatusFailedToProcess {
			return false, errTxUnknowCannotBeUpdated
		}
		return true, nil
	default:
		return false, errInvalidTxStatus
	}
}
