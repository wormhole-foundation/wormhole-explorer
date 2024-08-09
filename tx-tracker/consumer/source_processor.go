package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	notionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"

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
	TrackID     string
	Timestamp   *time.Time
	ChainId     sdk.ChainID
	ID          string // digest
	VaaId       string // {chain/address/sequence}
	Emitter     string
	Sequence    string
	TxHash      string
	Vaa         []byte
	IsVaaSigned bool
	Source      string
	// Overwrite indicates whether to reprocess a VAA that has already been processed.
	//
	// In the context of backfilling, sometimes you want to overwrite old data (e.g.: because
	// the schema changed).
	// In the context of the service, you usually don't want to overwrite existing data
	// to avoid processing the same VAA twice, which would result in performance degradation.
	Overwrite       bool
	Metrics         metrics.Metrics
	SentTimestamp   *time.Time
	DisableDBUpsert bool
	P2pNetwork      string
	DbLayer         config.DbLayer
}

func ProcessSourceTx(
	ctx context.Context,
	logger *zap.Logger,
	rpcPool map[vaa.ChainID]*pool.Pool,
	wormchainRpcPool map[vaa.ChainID]*pool.Pool,
	mongoRepository MongoDBRepository,
	params *ProcessSourceTxParams,
	p2pNetwork string,
	notionalCache *notionalCache.NotionalCache,
	sqlRepository PostgreSQLRepository,
) (*chains.TxDetail, error) {

	if !params.Overwrite {
		// If the message has already been processed, skip it.
		//
		// Sometimes the SQS visibility timeout expires and the message is put back into the queue,
		// even if the RPC nodes have been hit and data has been written to MongoDB.
		// In those cases, when we fetch the message for the second time,
		// we don't want to hit the RPC nodes again for performance reasons.
		var processedSQL bool
		if params.DbLayer != config.DbLayerMongo { // config.DbLayerPostgresql or config.DbLayerDual
			var errSQL error
			processedSQL, errSQL = sqlRepository.AlreadyProcessed(ctx, params.ID)
			if errSQL != nil {
				return nil, errSQL
			}
		}

		if params.DbLayer == config.DbLayerPostgresql {
			if processedSQL {
				return nil, ErrAlreadyProcessed
			}
		}

		var processedMongoDB bool
		if params.DbLayer != config.DbLayerPostgresql { // config.DbLayerMongo or config.DbLayerDual
			var errMongoDB error
			processedMongoDB, errMongoDB = mongoRepository.AlreadyProcessed(ctx, params.VaaId)
			if errMongoDB != nil {
				return nil, errMongoDB
			}
		}

		if params.DbLayer == config.DbLayerMongo {
			if processedMongoDB {
				return nil, ErrAlreadyProcessed
			}
		}

		if params.DbLayer == config.DbLayerDual {
			if processedSQL && processedMongoDB {
				return nil, ErrAlreadyProcessed
			}
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

	if params.IsVaaSigned && params.TxHash == "" {
		// add metrics for vaa without txHash
		params.Metrics.IncVaaWithoutTxHash(uint16(params.ChainId), params.Source)

		vaa, err := sdk.Unmarshal(params.Vaa)
		if err != nil {
			logger.Error("Error unmarshalling vaa", zap.Error(err), zap.String("vaaId", params.VaaId))
			return nil, errors.New("txHash is empty")
		}
		uniqueVaaID := domain.CreateUniqueVaaID(vaa)
		v, err := mongoRepository.GetVaaIdTxHash(ctx, uniqueVaaID)
		if err != nil {
			logger.Error("failed to find vaaIdTxHash",
				zap.String("trackId", params.TrackID),
				zap.String("vaaId", params.VaaId),
				zap.Any("vaaTimestamp", params.Timestamp),
				zap.Error(err),
			)
		} else {
			// add metrics for vaa with txHash fixed
			params.Metrics.IncVaaWithTxHashFixed(uint16(params.ChainId), params.Source)
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
	txDetail, err = chains.FetchTx(ctx, rpcPool, wormchainRpcPool, params.ChainId, params.TxHash, params.Timestamp, p2pNetwork, params.Metrics, logger, notionalCache)
	if err != nil {
		errHandleFetchTx := handleFetchTxError(ctx, logger, mongoRepository, params, err)
		if errHandleFetchTx == nil {
			params.Metrics.IncStoreUnprocessedOriginTx(uint16(params.ChainId))
		}
		return nil, err
	}

	// If disableDBUpsert is set to true, we don't want to store the source transaction details in the database.
	if params.DisableDBUpsert {
		return txDetail, nil
	}

	// Store source transaction details in the database
	p := UpsertOriginTxParams{
		Id:        params.ID,
		VaaId:     params.VaaId,
		TrackID:   params.TrackID,
		ChainId:   params.ChainId,
		Timestamp: params.Timestamp,
		TxDetail:  txDetail,
		TxStatus:  domain.SourceTxStatusConfirmed,
		Processed: true,
	}

	if params.DbLayer == config.DbLayerMongo || params.DbLayer == config.DbLayerDual {
		err = mongoRepository.UpsertOriginTx(ctx, &p)
		if err == nil {
			params.Metrics.VaaProcessingDuration(params.ChainId.String(), params.SentTimestamp)
		}
	}
	var errSQL error
	if params.DbLayer == config.DbLayerPostgresql || params.DbLayer == config.DbLayerDual {
		errSQL = upsertOriginTxPostresql(ctx, logger, sqlRepository, p, params, txDetail)
	}

	return txDetail, errors.Join(err, errSQL)
}

func upsertOriginTxPostresql(ctx context.Context, logger *zap.Logger, sqlRepository PostgreSQLRepository, p UpsertOriginTxParams, params *ProcessSourceTxParams, txDetail *chains.TxDetail) error {
	err := sqlRepository.UpsertOriginTx(ctx, &p)
	if err != nil {
		return err
	}

	if p.TxDetail.Attribute != nil && p.TxDetail.Attribute.Type == "wormchain-gateway" {
		if p.TxDetail.Attribute.Value == nil {
			logger.Error("wormchain attribute value is nil.", zap.String("vaaId", params.VaaId), zap.String("txHash", params.TxHash))
			return fmt.Errorf("failed to get wormchain attribute value. vaaId:%s - txHash:%s", params.VaaId, params.TxHash)
		}
		attr, ok := p.TxDetail.Attribute.Value.(*chains.WorchainAttributeTxDetail)
		if !ok {
			logger.Error("failed to convert to WorchainAttributeTxDetail", zap.String("vaaId", params.VaaId))
			return errors.New("failed to convert to WorchainAttributeTxDetail. vaaId: " + params.VaaId)
		}

		p = UpsertOriginTxParams{
			Id:        params.ID,
			VaaId:     params.VaaId,
			TrackID:   params.TrackID,
			ChainId:   attr.OriginChainID,
			Timestamp: params.Timestamp,
			TxDetail: &chains.TxDetail{
				From:         attr.OriginAddress,
				NativeTxHash: domain.NormalizeTxHashByChainId(attr.OriginChainID, attr.OriginTxHash),
				FeeDetail:    txDetail.FeeDetail,
			},
			TxStatus:  domain.SourceTxStatusConfirmed,
			Processed: true,
		}
		err = sqlRepository.UpsertOriginTx(ctx, &p)
		if err != nil {
			logger.Error("failed to upsert originTx-wormchain",
				zap.Error(err),
				zap.String("vaaId", params.VaaId))
			return err
		}
	}

	return sqlRepository.RegisterProcessedVaa(ctx, params.ID, params.VaaId)
}

func handleFetchTxError(ctx context.Context, logger *zap.Logger, repository MongoDBRepository, params *ProcessSourceTxParams, err error) error {
	// If the chain is not supported, we don't want to store the unprocessed originTx in the database.
	if errors.Is(chains.ErrChainNotSupported, err) {
		return nil
	}

	// if the transactions is solana or aptos, we don't want to store the txHash in the
	// unprocessed originTx in the database.
	var vaaTxDetail *chains.TxDetail
	isSolanaOrAptos := params.ChainId == vaa.ChainIDAptos || params.ChainId == vaa.ChainIDSolana
	if !isSolanaOrAptos {
		txHash := chains.FormatTxHashByChain(params.ChainId, params.TxHash)
		vaaTxDetail = &chains.TxDetail{
			NativeTxHash: txHash,
		}
	}

	e := UpsertOriginTxParams{
		Id:        params.ID,
		VaaId:     params.VaaId,
		TrackID:   params.TrackID,
		ChainId:   params.ChainId,
		Timestamp: params.Timestamp,
		TxDetail:  vaaTxDetail,
		TxStatus:  domain.SourceTxStatusConfirmed,
		Processed: false,
	}

	errUpsert := repository.UpsertOriginTx(ctx, &e)
	if errUpsert != nil {
		logger.Error("failed to upsert originTx",
			zap.Error(errUpsert),
			zap.String("vaaId", params.VaaId))
	}

	return nil
}
