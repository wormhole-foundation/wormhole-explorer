package consumer

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var (
	errTxFailedCannotBeUpdated = errors.New("tx with status failed can not be updated because exists a confirmed tx for the same vaa ID")
	errTxUnknowCannotBeUpdated = errors.New("tx with status unknown can not be updated because exists a tx (confirmed|failed) for the same vaa ID")
	errInvalidTxStatus         = errors.New("invalid tx status")
)

// ProcessTargetTxParams is a struct that contains the parameters for the ProcessTargetTx method.
type ProcessTargetTxParams struct {
	Source         string
	TrackID        string
	ID             string // digest
	VaaID          string // {chainID/address/sequence}
	ChainID        sdk.ChainID
	Emitter        string
	TxHash         string
	BlockTimestamp *time.Time
	BlockHeight    string
	Method         string
	From           string
	To             string
	Status         string
	EvmFee         *EvmFee
	SolanaFee      *SolanaFee
	Metrics        metrics.Metrics
	P2pNetwork     string
	RunMode        config.RunMode
}

type EvmFee struct {
	GasUsed           string
	EffectiveGasPrice string
}

type SolanaFee struct {
	Fee uint64
}

func ProcessTargetTx(
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
	params *ProcessTargetTxParams,
	notionalCache *notional.NotionalCache,
	postreSQLRepository PostgreSQLRepository,
) error {

	feeDetail := calculateFeeDetail(params, logger, notionalCache)

	txHash := domain.NormalizeTxHashByChainId(params.ChainID, params.TxHash)
	now := time.Now()
	update := &TargetTxUpdate{
		ID:      params.ID,
		VaaID:   params.VaaID,
		TrackID: params.TrackID,
		Destination: &DestinationTx{
			ChainID:     params.ChainID,
			Status:      params.Status,
			TxHash:      txHash,
			BlockNumber: params.BlockHeight,
			Timestamp:   params.BlockTimestamp,
			From:        params.From,
			To:          params.To,
			Method:      params.Method,
			FeeDetail:   feeDetail,
			UpdatedAt:   &now,
		},
	}

	// check if the transaction should be updated.
	shoudBeUpdated, err := checkTxShouldBeUpdated(ctx, update, repository)
	if !shoudBeUpdated {
		logger.Warn("Transaction should not be updated", zap.String("vaaId", params.VaaID), zap.Error(err))
		return nil
	}

	if params.RunMode == config.RunModeMongo || params.RunMode == config.RunModeDual {
		err = repository.UpsertTargetTx(ctx, update)
		if err == nil {
			params.Metrics.IncDestinationTxInserted(params.ChainID.String(), params.Source)
		}
	}

	var errSQL error
	if params.RunMode == config.RunModePostgres || params.RunMode == config.RunModeDual {
		errSQL = postreSQLRepository.UpsertTargetTx(ctx, update)
		if errSQL != nil {
			logger.Error("Error upserting target tx", zap.Error(errSQL), zap.String("vaaId", params.VaaID))
		}
	}
	return errors.Join(err, errSQL)
}

// Add an interface layer for the repository in order to decouple it from postresql and mongodb.
type getTxStatus interface {
	GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error)
}

func checkTxShouldBeUpdated(ctx context.Context, tx *TargetTxUpdate, repository getTxStatus) (bool, error) {
	switch tx.Destination.Status {
	case domain.DstTxStatusConfirmed:
		return true, nil
	case domain.DstTxStatusFailedToProcess:
		// check if the transaction exists from the same vaa ID.
		status, err := repository.GetTxStatus(ctx, tx)
		if err != nil {
			return true, nil
		}
		// if the transaction was already confirmed, then no update it.
		if status == domain.DstTxStatusConfirmed {
			return false, errTxFailedCannotBeUpdated
		}
		return true, nil
	case domain.DstTxStatusUnkonwn:
		// check if the transaction exists from the same vaa ID.
		status, err := repository.GetTxStatus(ctx, tx)
		if err != nil {
			return true, nil
		}
		// if the transaction was already confirmed or failed to process, then no update it.
		if status == domain.DstTxStatusConfirmed || status == domain.DstTxStatusFailedToProcess {
			return false, errTxUnknowCannotBeUpdated
		}
		return true, nil
	default:
		return false, errInvalidTxStatus
	}
}

func calculateFeeDetail(params *ProcessTargetTxParams, logger *zap.Logger, notionalCache *notional.NotionalCache) *FeeDetail {

	// calculate tx fee for evm redeemed tx.
	var feeDetail *FeeDetail
	if params.EvmFee != nil {
		fee, err := chains.EvmCalculateFee(params.ChainID, params.EvmFee.GasUsed, params.EvmFee.EffectiveGasPrice)
		if err != nil {
			logger.Error("can not calculated fee for redeemed tx",
				zap.Error(err),
				zap.String("txHash", params.TxHash),
				zap.String("chainId", params.ChainID.String()),
				zap.String("gasUsed", params.EvmFee.GasUsed),
				zap.String("effectiveGasPrice", params.EvmFee.EffectiveGasPrice),
			)
			return nil
		}
		if fee != nil {
			feeDetail = &FeeDetail{
				RawFee: map[string]string{
					"gasUsed":           params.EvmFee.GasUsed,
					"effectiveGasPrice": params.EvmFee.EffectiveGasPrice,
				},
				Fee: fee.String(),
			}
		}
	}
	// calculate tx fee for solana redeemed tx.
	if params.SolanaFee != nil {
		fee := chains.SolanaCalculateFee(params.SolanaFee.Fee)
		feeDetail = &FeeDetail{
			RawFee: map[string]string{
				"fee": strconv.FormatUint(params.SolanaFee.Fee, 10),
			},
			Fee: fee,
		}
	}

	if feeDetail != nil && params.P2pNetwork == domain.P2pMainNet {
		gasTokenPrice, errGasPrice := chains.GetGasTokenNotional(params.ChainID, notionalCache)
		if errGasPrice != nil {
			logger.Error("Failed to get gas price",
				zap.Error(errGasPrice),
				zap.String("chainId", params.ChainID.String()),
				zap.String("txHash", params.TxHash),
			)
			return feeDetail
		}
		feeDetail.GasTokenNotional = gasTokenPrice.NotionalUsd.String()
		feeDetail.FeeUSD = gasTokenPrice.NotionalUsd.Mul(decimal.RequireFromString(feeDetail.Fee)).String()
	}

	return feeDetail
}
